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
	"log"
	"strings"

	"golang.org/x/net/html/atom"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func iCalEscapeText(s string) string {
	var buf strings.Builder
	buf.Grow(len(s) * 2)
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\n':
			buf.WriteByte('\\')
			buf.WriteByte('n')
		case ',', ';', '\\':
			buf.WriteByte('\\')
			buf.WriteByte(c)
		default:
			buf.WriteByte(c)
		}
	}
	return buf.String()
}

func iCalEscapeParameterValue(s string) string {
	var buf strings.Builder
	buf.Grow(len(s) * 2)
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\n':
			buf.WriteByte('^')
			buf.WriteByte('n')
		case '"':
			buf.WriteByte('^')
			buf.WriteByte('\'')
		case '^':
			buf.WriteByte('^')
			buf.WriteByte('^')
		default:
			buf.WriteByte(c)
		}
	}
	return buf.String()
}

func innerText(s *goquery.Selection) string {
	var buf bytes.Buffer

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(strings.ReplaceAll(strings.TrimSpace(matchSpace.ReplaceAllString(n.Data, " ")), "\u00a0", " "))
		} else if n.Type == html.ElementNode && n.DataAtom == atom.Br {
			buf.WriteByte('\n')
		}
		if n.FirstChild != nil {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		if n.Type == html.ElementNode && (n.DataAtom == atom.Div || n.DataAtom == atom.P) {
			buf.WriteByte('\n')
		}
	}
	for _, n := range s.Nodes {
		f(n)
	}

	log.Printf("%q\n", buf.String())
	return buf.String()
}
