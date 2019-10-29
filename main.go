/*
  HKUST-MATH-Seminar-to-iCal -- Convert HKUST MATH department seminar
  calendar into iCal format
  Copyright (C) 2019  StarBrilliant

  This program is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/handlers"
	"golang.org/x/xerrors"
)

type handler struct {
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	iCal := bytes.NewBuffer(nil)
	iCalHasError := false
	status := 200
	wr := newICalWriter(iCal)

	err := iCalWriteHeader(wr)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = iCalWriteFooter(wr)
		if err != nil {
			panic(err)
		}
		if !iCalHasError {
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}
		w.Header().Set("Content-Length", strconv.FormatInt(int64(iCal.Len()), 10))
		w.Header().Set("Content-Type", "text/calendar; charset=UTF-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(status)
		w.Write(iCal.Bytes())
	}()

	crawlReq, err := http.NewRequestWithContext(r.Context(), "GET", "https://www.math.ust.hk/events/", nil)
	if err != nil {
		iCalWriteErrorMessage(wr, "Internal server error")
		iCalHasError = true
		status = http.StatusInternalServerError
		panic(err)
	}
	crawlReq.Header.Set("User-Agent", "Mozilla/5.0 HKUST-MATH-Seminar-to-iCal (+https://github.com/m13253/HKUST-MATH-Seminar-to-iCal)")
	resp, err := http.DefaultClient.Do(crawlReq)
	if err != nil {
		iCalWriteErrorMessage(wr, fmt.Sprintf("Failed to access www.math.ust.hk: %s", err))
		iCalHasError = true
		status = http.StatusBadGateway
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		iCalWriteErrorMessage(wr, fmt.Sprintf("Failed to access www.math.ust.hk: %s", resp.Status))
		iCalHasError = true
		status = http.StatusBadGateway
		return
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		iCalWriteErrorMessage(wr, fmt.Sprintf("Failed to read calendar data from www.math.ust.hk: %s", err))
		iCalHasError = true
		status = http.StatusBadGateway
		return
	}

	upcoming := doc.Find("b + table[width=\"100%\"][border=\"0\"][cellpadding=\"2\"][cellspacing=\"2\"]")
	if upcoming.Length() != 1 {
		iCalWriteErrorMessage(wr, "Failed to parse www.math.ust.hk: cannot locate the upcoming event list")
		iCalHasError = true
		status = http.StatusInternalServerError
		return
	}

	hkt, err := time.LoadLocation("Asia/Hong_Kong")
	if err != nil {
		iCalWriteErrorMessage(wr, "Internal server error")
		iCalHasError = true
		status = http.StatusInternalServerError
		panic(err)
	}
	now := time.Now().UTC()

	events := upcoming.Find("tr:nth-of-type(n+2)")
	events.Each(func(i int, s *goquery.Selection) {
		date := innerText(s.Find("td:nth-of-type(1)"))
		venue := innerText(s.Find("td:nth-of-type(2)"))
		title := innerText(s.Find("td:nth-of-type(3)"))
		attachment, _ := s.Find("td:nth-of-type(3) a").Attr("href")
		speaker := innerText(s.Find("td:nth-of-type(4)"))

		dateLines := strings.SplitN(date, "\n", 3)
		titleLines := strings.SplitN(title, "\n", 2)

		var summary string
		if len(titleLines) >= 2 {
			summary = titleLines[0] + ": " + titleLines[1]
		} else {
			summary = titleLines[0]
		}

		if len(dateLines) < 2 {
			if !iCalHasError {
				iCalWriteErrorMessage(wr, "Failed to parse the event:\n"+summary)
				iCalHasError = true
			}
			return
		}

		dateParsed := matchDate.FindStringSubmatch(strings.ToLower(dateLines[0]))
		if len(dateParsed) == 0 {
			if !iCalHasError {
				iCalWriteErrorMessage(wr, "Failed to parse the event date:\n"+summary)
				iCalHasError = true
			}
			return
		}
		timeParsed := matchTime.FindStringSubmatch(strings.ToLower(dateLines[1]))
		if len(timeParsed) == 0 {
			if !iCalHasError {
				iCalWriteErrorMessage(wr, "Failed to parse the event time:\n"+summary)
				iCalHasError = true
			}
			return
		}

		day, err := strconv.ParseInt(dateParsed[1], 10, 0)
		if err != nil {
			if !iCalHasError {
				iCalWriteErrorMessage(wr, "Failed to parse the day of the event date:\n"+summary)
				iCalHasError = true
			}
			return
		}
		month, ok := monthMap[dateParsed[2]]
		if !ok {
			if !iCalHasError {
				iCalWriteErrorMessage(wr, "Failed to parse the month of the event date:\n"+summary)
				iCalHasError = true
			}
			return
		}
		year, err := strconv.ParseInt(dateParsed[3], 10, 0)
		if err != nil {
			if !iCalHasError {
				iCalWriteErrorMessage(wr, "Failed to parse the year of the event date:\n"+summary)
				iCalHasError = true
			}
			return
		}
		starthour, err := strconv.ParseInt(timeParsed[1], 10, 0)
		if err != nil {
			if !iCalHasError {
				iCalWriteErrorMessage(wr, "Failed to parse the hour of the event start time:\n"+summary)
				iCalHasError = true
			}
			return
		}
		startmin, err := strconv.ParseInt(timeParsed[2], 10, 0)
		if err != nil {
			if !iCalHasError {
				iCalWriteErrorMessage(wr, "Failed to parse the minute of the event start time:\n"+summary)
				iCalHasError = true
			}
			return
		}
		startampm, ok := ampmMap[timeParsed[3]]
		if !ok {
			if !iCalHasError {
				iCalWriteErrorMessage(wr, "Failed to parse the a.m./p.m. part of the event start time:\n"+summary)
				iCalHasError = true
			}
			return
		}
		endhour, err := strconv.ParseInt(timeParsed[4], 10, 0)
		if err != nil {
			if !iCalHasError {
				iCalWriteErrorMessage(wr, "Failed to parse the hour of the event end time:\n"+summary)
				iCalHasError = true
			}
			return
		}
		endmin, err := strconv.ParseInt(timeParsed[5], 10, 0)
		if err != nil {
			if !iCalHasError {
				iCalWriteErrorMessage(wr, "Failed to parse the minute of the event end time:\n"+summary)
				iCalHasError = true
			}
			return
		}
		endampm, ok := ampmMap[timeParsed[6]]
		if !ok {
			if !iCalHasError {
				iCalWriteErrorMessage(wr, "Failed to parse the a.m./p.m. part of the event end time:\n"+summary)
				iCalHasError = true
			}
			return
		}

		dtstart := time.Date(int(year), month, int(day), int(starthour)%12+startampm, int(startmin), 0, 0, hkt)
		dtend := time.Date(int(year), month, int(day), int(endhour)%12+endampm, int(endmin), 0, 0, hkt)

		_, err = io.WriteString(wr, "BEGIN:VEVENT\n")
		if err != nil {
			panic(xerrors.Errorf("failed to write calendar event: %w", err))
		}
		_, err = io.WriteString(wr, fmt.Sprintf("DTSTART;TZID=Asia/Hong_Kong:%s\n", dtstart.Format("20060102T150405")))
		if err != nil {
			panic(xerrors.Errorf("failed to write calendar event: %w", err))
		}
		_, err = io.WriteString(wr, fmt.Sprintf("DTEND;TZID=Asia/Hong_Kong:%s\n", dtend.Format("20060102T150405")))
		if err != nil {
			panic(xerrors.Errorf("failed to write calendar event: %w", err))
		}
		_, err = io.WriteString(wr, fmt.Sprintf("DTSTAMP:%s\n", now.Format("20060102T150405Z")))
		if err != nil {
			panic(xerrors.Errorf("failed to write calendar event: %w", err))
		}
		_, err = io.WriteString(wr, fmt.Sprintf("DESCRIPTION:%s\n", iCalEscapeText(title)))
		if err != nil {
			panic(xerrors.Errorf("failed to write calendar event: %w", err))
		}
		_, err = io.WriteString(wr, fmt.Sprintf("LAST-MODIFIED:%s\n", now.Format("20060102T150405Z")))
		if err != nil {
			panic(xerrors.Errorf("failed to write calendar event: %w", err))
		}
		_, err = io.WriteString(wr, fmt.Sprintf("LOCATION:%s\n", iCalEscapeText(venue)))
		if err != nil {
			panic(xerrors.Errorf("failed to write calendar event: %w", err))
		}
		_, err = io.WriteString(wr, fmt.Sprintf("ORGANIZER:%s\n", iCalEscapeText(speaker)))
		if err != nil {
			panic(xerrors.Errorf("failed to write calendar event: %w", err))
		}
		_, err = io.WriteString(wr, "SEQUENCE:0\n")
		if err != nil {
			panic(xerrors.Errorf("failed to write calendar event: %w", err))
		}
		_, err = io.WriteString(wr, fmt.Sprintf("SUMMARY:%s\n", iCalEscapeText(summary)))
		if err != nil {
			panic(xerrors.Errorf("failed to write calendar event: %w", err))
		}
		if len(attachment) != 0 {
			_, err = io.WriteString(wr, fmt.Sprintf("URI:%s\n", iCalEscapeText(attachment)))
			if err != nil {
				panic(xerrors.Errorf("failed to write calendar event: %w", err))
			}
		}
		_, err = io.WriteString(wr, "END:VEVENT\n")
		if err != nil {
			panic(xerrors.Errorf("failed to write calendar event: %w", err))
		}
	})
}

func main() {
	h := &handler{}

	servemux := http.NewServeMux()
	servemux.Handle("/math-seminar.ics", h)
	servemux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/math-seminar.ics", http.StatusFound)
	})
	log.Println("Listening on http://*:19777/math-seminar.ics")
	http.ListenAndServe(":19777", handlers.CombinedLoggingHandler(os.Stdout, servemux))
}
