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
	"regexp"
	"time"
)

var (
	matchSpace = regexp.MustCompile(`\s+`)
	matchDate  = regexp.MustCompile(`^\s*(\d+)\s+(jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)\s+(\d+)\s+[a-z]+\s*$`)
	matchTime  = regexp.MustCompile(`^\s*\(\s*(\d+):(\d+)\s*(am|a\.m\.|pm|p\.m\.)\s*-\s*(\d+):(\d+)\s*(am|a\.m\.|midnight|night|pm|p\.m\.|noon)\s*\)\s*$`)
	monthMap   = map[string]time.Month{
		"jan": time.January,
		"feb": time.February,
		"mar": time.March,
		"apr": time.April,
		"may": time.May,
		"jun": time.June,
		"jul": time.July,
		"aug": time.August,
		"sep": time.September,
		"oct": time.October,
		"nov": time.November,
		"dec": time.December,
	}
	ampmMap = map[string]int{
		"am":       0,
		"a.m.":     0,
		"midnight": 0,
		"night":    0,
		"pm":       12,
		"p.m.":     12,
		"noon":     12,
	}
)
