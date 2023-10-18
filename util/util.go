package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
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
		// log.Fatalf("Could not load location: %v\n", err)
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

// Gets the line count of a file
func GetLineCount(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}


// Returns a ICalTimestamp string by an input date time
func TimeToICalTimestamp(t *time.Time) (string, error) {
	year := PadInt(t.Year(), 2)
	month := PadInt(int(t.Month()), 2)
	day := PadInt(t.Day(), 2)
	hour := PadInt(t.Hour(), 2)
	minute := PadInt(t.Minute(), 2)

	str := fmt.Sprintf("%s%s%sT%s%s00Z", year, month, day, hour, minute)
	return str, nil
}


// Returns a string representation of an integer with given amount of padding zeroes
func PadInt(i int, count int) string {
	layout := fmt.Sprintf("%%0%dd", count)
	fmt.Println(layout)
	return fmt.Sprintf(layout, i)
}

// Returns a date time object given a ICalTimestamp string
func ICalTimestampToTime (stamp string) (*time.Time, error) {
	// "0 1 2 3 | 45 | 67 | 8 | 9 10 | 11 12 | 13 14 | 15"
	// "1 9 9 7 | 07 | 15 | T | 0 4  | 0  0  | 0  0  | Z"

	year, err := strconv.Atoi(stamp[:4])
	if err != nil {
		return &time.Time{}, err
	}
	month, err := strconv.Atoi(stamp[4:6])
	if err != nil {
		return &time.Time{}, err
	}
	day, err := strconv.Atoi(stamp[6:8])
	if err != nil {
		return &time.Time{}, err
	}
	hour, err := strconv.Atoi(stamp[9:11])
	if err != nil {
		return &time.Time{}, err
	}
	minute, err := strconv.Atoi(stamp[11:13])
	if err != nil {
		return &time.Time{}, err
	}
	second, err := strconv.Atoi(stamp[13:15])
	if err != nil {
		return &time.Time{}, err
	}

	location, err := time.LoadLocation("Europe/Copenhagen")
	date := time.Date(year, time.Month(month), day, hour, minute, second, 0, location)
	return &date, err
}
