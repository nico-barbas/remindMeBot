package main

import (
	"strconv"
	"time"
)

const (
	bellEmote        = ":bell:"
	todoEmote        = ":clipboard:"
	todoCheckEmote   = ":white_check_mark:"
	todoUncheckEmote = ":negative_squared_cross_mark:"
)

type date struct {
	hour  int
	min   int
	day   int
	month time.Month
	year  int
}

func makeDate(ddmmyy [3]token, hhmm [2]token) date {
	result := date{}

	y, m, d := time.Now().Date()
	if ddmmyy[0].kind == tokenInvalid {
		result.day = d
	} else {
		day, _ := strconv.ParseInt(ddmmyy[0].text, 0, 64)
		result.day = int(day)
	}

	if ddmmyy[1].kind == tokenInvalid {
		result.month = m
	} else {
		month, _ := strconv.ParseInt(ddmmyy[1].text, 0, 64)
		result.month = time.Month(month)
	}

	if ddmmyy[2].kind == tokenInvalid {
		result.year = y
	} else {
		yearStr := "20" + ddmmyy[2].text
		year, _ := strconv.ParseInt(yearStr, 0, 64)
		result.year = int(year)
	}

	if hhmm[0].kind == tokenInvalid {
		result.hour = 0
	} else {
		hour, _ := strconv.ParseInt(hhmm[0].text, 0, 64)
		result.hour = int(hour)
	}

	if hhmm[1].kind == tokenInvalid {
		result.min = 0
	} else {
		min, _ := strconv.ParseInt(hhmm[1].text, 0, 64)
		result.min = int(min)
	}
	return result
}

func (self date) isEqual(d date) bool {
	if self.day != d.day {
		return false
	}
	if self.month != d.month {
		return false
	}
	if self.year != d.year {
		return false
	}
	if self.hour != d.hour {
		return false
	}
	if self.min != d.min {
		return false
	}
	return true
}

func findItemByName(buf []item, name string) int {
	for i := range buf {
		if buf[i].name == name {
			return i
		}
	}
	return -1
}
