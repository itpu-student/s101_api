package utils

import (
	"time"

	"github.com/itpu-student/s101_api/models"
)

// IsOpen returns whether a place is currently open given its weekly_hours
// and a reference time. Comparison happens in the supplied time's local zone.
// Ranges that cross midnight (e.g. 22:00 -> 02:00) are supported.
func IsOpen(wh models.WeeklyHours, now time.Time) *bool {
	if wh.IsBlank() {
		return nil
	}

	day := dayRanges(wh, now.Weekday())
	prev := dayRanges(wh, (now.Weekday()+6)%7) // yesterday
	cur := hhmm(now)
	pt, pf := true, false

	for _, r := range day {
		if isWithin(r, cur, false) {
			return &pt
		}
	}
	// yesterday's overnight range that extended past midnight
	for _, r := range prev {
		if r.Close < r.Open && cur < r.Close {
			return &pt
		}
	}
	return &pf
}

func isWithin(r models.HourRange, cur string, overnight bool) bool {
	if r.Open == "" || r.Close == "" {
		return false
	}
	if r.Open <= r.Close {
		return cur >= r.Open && cur < r.Close
	}
	// overnight range — from r.Open to end of day
	return cur >= r.Open
}

func hhmm(t time.Time) string {
	return t.Format("15:04")
}

func dayRanges(wh models.WeeklyHours, d time.Weekday) []models.HourRange {
	switch d {
	case time.Monday:
		return wh.Mon
	case time.Tuesday:
		return wh.Tue
	case time.Wednesday:
		return wh.Wed
	case time.Thursday:
		return wh.Thu
	case time.Friday:
		return wh.Fri
	case time.Saturday:
		return wh.Sat
	case time.Sunday:
		return wh.Sun
	}
	return nil
}
