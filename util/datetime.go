package util

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

const Day = 24 * time.Hour

const ISO8601Layout = "2006-01-02T15:04:05-07:00"

func MaxTime(t1 time.Time, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t2
	}

	return t1
}

func NoCheckParseISO8601(s string) time.Time {
	layout := ISO8601Layout
	t, _ := time.Parse(layout, s)
	return t
}

func ParseISO8601Timestamp(s string) (time.Time, error) {
	layout := ISO8601Layout
	t, err := time.Parse(layout, s)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// Given a time.Time, truncates it to start of the date.
func TruncateToStartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func StartOfDay(year, month, day int, tz *time.Location) time.Time {
	if tz == nil {
		tz = time.UTC
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, tz)
}

// The endDate is exclusive i.e the range is [startDate .. endDate-1]. Also make
// sure that both are at 00:00 hours at the intended timezone, otherwise you get
// wrong results.
func DaysBetween(startDate, endDate time.Time) int {
	return int(endDate.Sub(startDate).Hours() / 24)
}

// SetTimezone takes the coordinates of t (coordinates being year, month, day,
// hr, min, sec, nsec) and creates a new time.Time with the same coordinates but
// timezone set to the given timezone. This does *not* return the same time
// represented by t relative to a different timezone, simply changes the
// timezone. This only makes sense when you're parsing time strings usint
// time.Parse and do not have timezone information in the string itself.
func setTimezone(t time.Time, timezone *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), timezone)
}

// Parses a date string of the form yyyy-mm-dd
func ParseDateString(date string, timezone *time.Location) (time.Time, error) {
	// Reference time: 01/02 03:04:05PM '06 -0700
	layout := "2006-01-02"
	t, err := time.Parse(layout, date)
	if err != nil {
		log.Printf("failed to parse date string %s: %s", date, err)
		log.Error().Err(err).Str("dateString", date).Msg("failed to parse date string")
		return time.Now(), err
	}
	return setTimezone(t, timezone), nil
}

func MustParseDateString(date string, timezone *time.Location) time.Time {
	// Reference time: 01/02 03:04:05PM '06 -0700
	layout := "2006-01-02"
	t, err := time.Parse(layout, date)
	if err != nil {
		log.Panic().Str("dateStr", date).Err(err).Msg("failed to parse date string")
		return time.Now()
	}
	return setTimezone(t, timezone)
}

func DateToYYYYMMDD(date time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d", date.Year(), date.Month(), date.Day())
}

func DateStringFromTime(t time.Time) string {
	return fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day()) // TODO: @Soumik, why not use time.format
}

// If day is monday, returns day. Otherwise returns the monday before day.
func TruncateToMonday(day time.Time) time.Time {
	dayOfWeek := day.Weekday()

	if dayOfWeek == time.Sunday {
		dayOfWeek = 7
	}

	daysSinceMonday := dayOfWeek - 1 // Since time.Monday == 1
	return day.Add(-24 * time.Hour * time.Duration(daysSinceMonday))
}

func TruncateToStartOfMonth(day time.Time) time.Time {
	return time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, day.Location())
}

func ExtendToStartOfNextMonth(day time.Time) time.Time {
	month := day.Month()
	for day.Month() == month {
		day = day.Add(24 * time.Hour)
	}
	return day
}

// Returns the date of previous week. weekStartsAt denotes the weekday on which
// we consider weeks to start at. For example, ISO-8601 convention wise, a week
// starts on monday.
func StartOfPreviousWeek(startDate time.Time, weekStartsAt time.Weekday) time.Time {
	startDate = TruncateToStartOfDay(startDate)
	weekOffset := int(startDate.Weekday())
	oneWeekBefore := startDate.Add(-24 * 7 * time.Hour)
	startOfPreviousWeek := oneWeekBefore.Add(-24 * time.Hour * time.Duration(weekOffset))
	return startOfPreviousWeek.Add(24 * time.Hour * time.Duration(weekStartsAt))
}
