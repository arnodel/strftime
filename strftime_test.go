package strftime

import (
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

var testTime = time.Date(2009, time.November, 10, 23, 1, 2, 3, time.UTC)
var testCases = []struct {
	format string
	value  string
}{
	{"%a", "Tue"},
	{"%A", "Tuesday"},
	{"%b", "Nov"},
	{"%B", "November"},
	{"%c", "Tue, 10 Nov 2009 23:01:02 UTC"},
	{"%d", "10"},
	{"%H", "23"},
	{"%I", "11"},
	{"%j", "314"},
	{"%m", "11"},
	{"%M", "01"},
	{"%p", "PM"},
	{"%S", "02"},
	{"%U", "45"},
	{"%w", "2"},
	{"%W", "45"},
	{"%x", "11/10/09"},
	{"%X", "23:01:02"},
	{"%y", "09"},
	{"%Y", "2009"},
	{"%Z", "UTC"},

	// Escape
	{"%%%Y", "%2009"},
	// Embedded
	{"/path/%Y/%m/report", "/path/2009/11/report"},
	//Empty
	{"", ""},
}

func TestFormats(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.format, func(t *testing.T) {
			value, err := Format(tc.format, testTime)
			if err != nil {
				t.Fatalf("error formatting %s - %s", tc.format, err)
			}
			if value != tc.value {
				t.Fatalf("got %s,  expected %s", value, tc.value)
			}
		})
	}
}

func TestUnknown(t *testing.T) {
	_, err := Format("%g", testTime)
	if err == nil {
		t.Fatal("managed to expand 'g'")
	}
}

func TestStrict(t *testing.T) {
	for _, s := range []string{"%", "%1", "%g", "%^", "% "} {
		t.Run(s, func(t *testing.T) {
			_, err := StrictFormat(s, testTime)
			if err == nil {
				t.Fatalf("managed to expand %q in strict mode", s)
			}
		})
	}
}

func TestQuick(t *testing.T) {
	fn := func(text string) bool {
		s, _ := Format(text, testTime)
		// TODO: Find better heuristic
		return len(s) >= len(text)
	}
	if err := quick.Check(fn, nil); err != nil {
		t.Fatal(err)
	}
}

func TestPadI(t *testing.T) {
	ts := time.Date(2020, 07, 27, 9, 8, 7, 0, time.UTC)
	out, err := Format("%I:%M:%S %p", ts)
	if err != nil {
		t.Fatal(err)
	}
	if "09:08:07 AM" != out {
		t.Fatal(out)
	}
}

// dayAndWeekOfYear computes %j, %U and %W from the calendar date alone (the
// year, month and day as written on the wall clock), without using the time's
// location: the day of the year starting at 1, and the week number counting
// weeks that start on Sunday (%U) or Monday (%W), with the days before the
// first such day in week 0.
func dayAndWeekOfYear(t time.Time) (yday, sunWeek, monWeek int) {
	date := time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.UTC)
	yday = date.YearDay()
	wday := int(date.Weekday())
	sunWeek = (yday - 1 + 7 - wday) / 7
	monWeek = (yday - 1 + 7 - (wday+6)%7) / 7
	return
}

// TestDayAndWeekOfYearInLocalTime checks %j, %U and %W for every day of two
// years (one a leap year), at three times of day, in zones west of UTC, east
// of UTC and in UTC. They are determined by the date on the wall clock, not
// by the offset of the zone.
func TestDayAndWeekOfYearInLocalTime(t *testing.T) {
	zones := []*time.Location{
		time.UTC,
		time.FixedZone("UTC-8", -8*3600),
		time.FixedZone("UTC+13", 13*3600),
	}
	for _, zone := range zones {
		for _, year := range []int{2020, 2021} {
			for day := time.Date(year, time.January, 1, 0, 0, 0, 0, zone); day.Year() == year; day = day.AddDate(0, 0, 1) {
				for _, hm := range [][2]int{{0, 0}, {12, 30}, {23, 59}} {
					ts := time.Date(day.Year(), day.Month(), day.Day(), hm[0], hm[1], 0, 0, zone)
					yday, sunWeek, monWeek := dayAndWeekOfYear(ts)
					for format, want := range map[string]int{"%j": yday, "%U": sunWeek, "%W": monWeek} {
						width := "%02d"
						if format == "%j" {
							width = "%03d"
						}
						got, err := Format(format, ts)
						if err != nil {
							t.Fatal(err)
						}
						if expected := fmt.Sprintf(width, want); got != expected {
							t.Fatalf("%s of %s: got %s, expected %s", format, ts.Format(time.RFC3339), got, expected)
						}
					}
				}
			}
		}
	}
}

// TestDayAndWeekOfYearValues pins literal values, taken from reference Lua
// 5.4.8 (os.date) for the same wall clock times, so the checks above do not
// rest only on the formulas the implementation uses.
func TestDayAndWeekOfYearValues(t *testing.T) {
	west := time.FixedZone("UTC-8", -8*3600)
	east := time.FixedZone("UTC+13", 13*3600)
	for _, tc := range []struct {
		when    time.Time
		j, U, W string
	}{
		// Before the first Sunday and the first Monday of the year: week 0.
		{time.Date(2020, 1, 1, 23, 59, 0, 0, time.UTC), "001", "00", "00"},
		{time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), "001", "00", "00"},
		// The first Sunday starts week 1 for %U, the first Monday for %W.
		{time.Date(2021, 1, 3, 12, 0, 0, 0, time.UTC), "003", "01", "00"},
		{time.Date(2021, 1, 4, 12, 0, 0, 0, time.UTC), "004", "01", "01"},
		// The last days of a leap year and of a common year.
		{time.Date(2020, 12, 31, 12, 0, 0, 0, time.UTC), "366", "52", "52"},
		{time.Date(2021, 12, 31, 23, 59, 0, 0, time.UTC), "365", "52", "52"},
		// The 29th of February.
		{time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC), "060", "08", "09"},
		// The same wall clock time in zones west and east of UTC.
		{time.Date(2020, 12, 31, 23, 0, 0, 0, west), "366", "52", "52"},
		{time.Date(2020, 1, 1, 23, 59, 0, 0, east), "001", "00", "00"},
		{time.Date(2020, 12, 31, 12, 0, 0, 0, east), "366", "52", "52"},
	} {
		for format, want := range map[string]string{"%j": tc.j, "%U": tc.U, "%W": tc.W} {
			got, err := Format(format, tc.when)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Errorf("%s of %s: got %s, expected %s", format, tc.when.Format(time.RFC3339), got, want)
			}
		}
	}
}
