package lectioutil

import (
	"strings"
	"time"
)

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
