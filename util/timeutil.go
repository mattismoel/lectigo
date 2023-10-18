package util

import "time"

// Returns the input date at midnight
func RoundDateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// Gets the date of the monday of the current week.
func GetMonday() (time.Time, error) {
	location, err := time.LoadLocation("Europe/Copenhagen")
	if err != nil {
		return time.Time{}, err
	}
	t := time.Now()
	off := int(t.Weekday()) - int(time.Monday)
	if off < 0 {
		off += 7 // Adjust if today is Sunday or earlier in the week
	}
	return time.Date(t.Year(), t.Month(), t.Day()-off, 0, 0, 0, 0, location), nil
}

