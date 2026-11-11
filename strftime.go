/*
Package strftime implements Python's strftime in Go

Example:
	str, err := strftime.Format("%Y/%m/%d", time.Now()) // 2019/07/07

Directives:
	%a - Locale’s abbreviated weekday name
	%A - Locale’s full weekday name
	%b - Locale’s abbreviated month name
	%B - Locale’s full month name
	%c - Locale’s appropriate date and time representation
	%d - Day of the month as a decimal number [01,31]
	%H - Hour (24-hour clock) as a decimal number [00,23]
	%I - Hour (12-hour clock) as a decimal number [01,12]
	%j - Day of the year as a decimal number [001,366]
	%m - Month as a decimal number [01,12]
	%M - Minute as a decimal number [00,59]
	%p - Locale’s equivalent of either AM or PM
	%S - Second as a decimal number [00,61]
	%U - Week number of the year [00,53], weeks start on Sunday, days before the first Sunday are in week 0
	%w - Weekday as a decimal number
	%W - Week number of the year [00,53], weeks start on Monday, days before the first Monday are in week 0
	%x - Locale’s appropriate date representation
	%X - Locale’s appropriate time representation
	%y - Year without century as a decimal number [00,99]
	%Y - Year with century as a decimal number
	%Z - Time zone name (no characters if no time zone exists)

Note that %c returns RFC1123 which is a bit different from what Python does
*/
package strftime

import (
	"fmt"
	"regexp"
	"time"
)

const (
	// Version is the package version
	Version = "0.1.5"

	// Week in time.Duration
	Week = time.Hour * 24 * 7
)

// See http://docs.python.org/3/library/time.html#time.strftime
var conv = map[string]string{
	"%%": "%",          // Literal '%'
	"%a": "Mon",        // Locale’s abbreviated weekday name
	"%A": "Monday",     // Locale’s full weekday name
	"%b": "Jan",        // Locale’s abbreviated month name
	"%B": "January",    // Locale’s full month name
	"%c": time.RFC1123, // Locale’s appropriate date and time representation
	"%d": "02",         // Day of the month as a decimal number [01,31]
	"%H": "15",         // Hour (24-hour clock) as a decimal number [00,23]
	"%I": "03",         // Hour (12-hour clock) as a decimal number [01,12]
	"%m": "01",         // Month as a decimal number [01,12]
	"%M": "04",         // Minute as a decimal number [00,59]
	"%p": "PM",         // Locale’s equivalent of either AM or PM
	"%S": "05",         // Second as a decimal number [00,61]
	"%x": "01/02/06",   // Locale’s appropriate date representation
	"%X": "15:04:05",   // Locale’s appropriate time representation
	"%y": "06",         // Year without century as a decimal number [00,99]
	"%Y": "2006",       // Year with century as a decimal number
	"%Z": "MST",        // Time zone name (no characters if no time zone exists)
}

var (
	fmtRe       = regexp.MustCompile("%[%a-zA-Z]")
	strictFmtRe = regexp.MustCompile("%.?")
)

// repl replaces % directives with right time, will panic on unknown directive
func repl(match string, t time.Time) (string, error) {
	format, ok := conv[match]
	if ok {
		return t.Format(format), nil
	}

	switch match {
	case "%j":
		return fmt.Sprintf("%03d", t.YearDay()), nil
	case "%w":
		return fmt.Sprintf("%d", t.Weekday()), nil
	case "%U":
		// Weeks start on Sunday; days before the first Sunday are in week 0.
		return fmt.Sprintf("%02d", (t.YearDay()-1+7-int(t.Weekday()))/7), nil
	case "%W":
		// Weeks start on Monday; days before the first Monday are in week 0.
		return fmt.Sprintf("%02d", (t.YearDay()-1+7-(int(t.Weekday())+6)%7)/7), nil
	}

	return "", fmt.Errorf("unknown directive - %s", match)
}

func doFormat(fmtRe *regexp.Regexp, format string, t time.Time) (string, error) {
	var err error

	fn := func(match string) string {
		s, rerr := repl(match, t)
		if rerr != nil {
			err = rerr
		}
		return s
	}

	return fmtRe.ReplaceAllStringFunc(format, fn), err
}

// Format return string with % directives expanded.
// Will return error on unknown directive.
func Format(format string, t time.Time) (string, error) {
	return doFormat(fmtRe, format, t)
}

// StrictFormat return string with % directives expanded. Will return error on
// unknown directive and on '%' not followed by alphabetic characters (e.g.
// '%1')
func StrictFormat(format string, t time.Time) (string, error) {
	return doFormat(strictFmtRe, format, t)
}
