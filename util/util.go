package util

import (
	"encoding/json"
	"strings"
	"time"
)

// Creates a map consisting of all values from both input maps
func MergeMaps[K comparable, V any](m1 map[K]V, m2 map[K]V) map[K]V {
	merged := make(map[K]V)

	for key, value := range m1 {
		merged[key] = value
	}
	for key, value := range m2 {
		merged[key] = value
	}

	return merged
}

// Compares two maps A and B, and returns two maps consisting of extras and missing from A
func CompareMaps[K comparable, V any](from map[K]V, to map[K]V) (extras map[K]V, missing map[K]V) {
	extras = make(map[K]V)
	missing = make(map[K]V)

	// If key of m1 does not exist in m2, add to the missing map
	for key, value := range from {
		if _, exists := to[key]; !exists {
			missing[key] = value
			// fmt.Printf("ID %v does not exist in to-map\n", key)
		}
	}

	for key, value := range to {
		if _, exists := from[key]; !exists {
			extras[key] = value
			// fmt.Printf("ID %v does not exist in from-map and is extra\n", key)
		}
	}

	return extras, missing
}

// Returns the input date at midnight
func RoundDateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// Returns a JSON string representation of a struct
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
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

// Rerturns a Lectio status based on the color id of a Google Calendar event
func StatusFromColorID(colorId string) string {
	switch colorId {
	case "4":
		return "aflyst"
	case "2":
		return "ændret"
	}
	return "uændret"
}

// Returns a Google Calendar color ID from a Lectio module status
// Aflyst: "4" - red
// Ændret: "2" - green
// Default "" - default calendar color
func ColorIDFromStatus(status string) string {
	switch status {
	case "aflyst":
		return "4"
	case "ændret":
		return "2"
	}
	return ""
}

func ConvertLectioDate(s string) (startTime time.Time, endTime time.Time, err error) {
	location, err := time.LoadLocation("Europe/Copenhagen")
	if err != nil {
		return startTime, endTime, err
	}
	layout := "2/1-2006 15:04"
	split := strings.Split(s, " til ")
	if len(split) != 2 {
		return startTime, endTime, err
	}

	startTime, err = time.ParseInLocation(layout, split[0], location)
	if err != nil {
		return startTime, endTime, err
	}

	date := startTime.Format("2/1-2006")
	endTime, err = time.ParseInLocation(layout, date+" "+split[1], location)
	if err != nil {
		return startTime, endTime, err
	}

	return startTime, endTime, nil
}
