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
	"fmt"
	"io"
	"log"
	"time"

	"golang.org/x/xerrors"
)

type iCalWriter struct {
	wr  io.Writer
	col int
}

func newICalWriter(w io.Writer) *iCalWriter {
	return &iCalWriter{
		wr: w,
	}
}

func (w *iCalWriter) Write(s []byte) (n int, err error) {
	for i, c := range s {
		if c == '\n' {
			_, err = w.wr.Write([]byte{'\r', '\n'})
			if err != nil {
				return i, err
			}
			w.col = 0
			continue
		}
		threshold := 75
		switch {
		case c < 0xc0:
		case c < 0xe0:
			threshold = 74
		case c < 0xf0:
			threshold = 73
		default:
			threshold = 72
		}
		if w.col >= threshold {
			_, err = w.wr.Write([]byte{'\r', '\n', ' ', c})
			if err != nil {
				return i, err
			}
			w.col = 2
		} else {
			_, err = w.wr.Write([]byte{c})
			if err != nil {
				return i, err
			}
			w.col++
		}
	}
	return len(s), nil
}

func iCalWriteHeader(w *iCalWriter) error {
	const header = "BEGIN:VCALENDAR\n" +
		"PRODID:-//StarBrilliant//HKUST-MATH-Seminar-to-iCal//EN\n" +
		"VERSION:2.0\n" +
		"CALSCALE:GREGORIAN\n" +
		"METHOD:PUBLISH\n" +
		"REFRESH-INTERVAL;VALUE=DURATION:PT1H\n" +
		"X-PUBLISHED-TTL:PT1H\n" +
		"X-WR-CALNAME:HKUST MATH Seminars\n" +
		"X-WR-TIMEZONE:Asia/Hong_Kong\n" +
		"BEGIN:VTIMEZONE\n" +
		"TZID:Asia/Hong_Kong\n" +
		"X-LIC-LOCATION:Asia/Hong_Kong\n" +
		"BEGIN:STANDARD\n" +
		"TZOFFSETFROM:+0800\n" +
		"TZOFFSETTO:+0800\n" +
		"TZNAME:HKT\n" +
		"DTSTART:19700101T000000\n" +
		"END:STANDARD\n" +
		"END:VTIMEZONE\n"
	_, err := io.WriteString(w, header)
	if err != nil {
		return xerrors.Errorf("failed to write calendar: %w", err)
	}
	return nil
}

func iCalWriteFooter(w *iCalWriter) error {
	const footer = "END:VCALENDAR\n"
	_, err := io.WriteString(w, footer)
	if err != nil {
		return xerrors.Errorf("failed to write calendar: %w", err)
	}
	return nil
}

func iCalWriteErrorMessage(w *iCalWriter, msg string) {
	const header = "BEGIN:VEVENT\n"
	const footer = "SEQUENCE:0\n" +
		"STATUS:CONFIRMED\n" +
		"SUMMARY:HKUST calendar parser had a problem\n" +
		"TRANSP:TRANSPARENT\n" +
		"END:VEVENT\n"

	log.Println(msg)

	hkt, err := time.LoadLocation("Asia/Hong_Kong")
	if err != nil {
		panic(err)
	}
	now := time.Now()
	utc := now.UTC()
	local := now.In(hkt)

	_, err = io.WriteString(w, header)
	if err != nil {
		panic(xerrors.Errorf("failed to write error message: %w", err))
	}
	_, err = io.WriteString(w, fmt.Sprintf("DTSTART;VALUE=DATE:%s\n", local.Format("20060102")))
	if err != nil {
		panic(xerrors.Errorf("failed to write error message: %w", err))
	}
	_, err = io.WriteString(w, fmt.Sprintf("DTEND;VALUE=DATE:%s\n", local.Format("20060102")))
	if err != nil {
		panic(xerrors.Errorf("failed to write error message: %w", err))
	}
	_, err = io.WriteString(w, fmt.Sprintf("DTSTAMP:%s\n", utc.Format("20060102T150405Z")))
	if err != nil {
		panic(xerrors.Errorf("failed to write error message: %w", err))
	}
	_, err = io.WriteString(w, fmt.Sprintf("DESCRIPTION:%s\n", iCalEscapeText(msg)))
	if err != nil {
		panic(xerrors.Errorf("failed to write error message: %w", err))
	}
	_, err = io.WriteString(w, fmt.Sprintf("LAST-MODIFIED:%s\n", utc.Format("20060102T150405Z")))
	if err != nil {
		panic(xerrors.Errorf("failed to write error message: %w", err))
	}
	_, err = io.WriteString(w, footer)
	if err != nil {
		panic(xerrors.Errorf("failed to write error message: %w", err))
	}
}
