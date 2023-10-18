package types

import (
	"fmt"
	"os"
	"time"

	"github.com/mattismoel/lectigo/util"
)

type ICalendar struct {
	Coorperation string
	ProductName  string
	Events       []*ICalEvent
}

type ICalEvent struct {
	UID         string
	StartDate   time.Time
	EndDate     time.Time
	Summary     string
	Location    string
	Description string
}

// Returns the string representation of an ICalendar struct. This can be input into an .ics file.
func (e *ICalEvent) ToString() (string, error) {
	layout := `BEGIN:VEVENT
UID:%s
DTSTAMP:%s
ORGANIZER;CN=John Doe:MAILTO:john.doe@example.com
DTSTART:%s
DTEND:%s
SUMMARY:%s 
LOCATION:%s
END:VEVENT
`

	startDate, err := util.TimeToICalTimestamp(&e.StartDate)
	endDate, err := util.TimeToICalTimestamp(&e.EndDate)
	if err != nil {
		return "", err
	}
	eventStr := fmt.Sprintf(layout, e.UID, startDate, startDate, endDate, e.Location)
	return eventStr, nil
}

// Writes all events of an ICalendar to a specified file path. This function must be called everytime
// one wishes to update an ICalendar file with new events.
func (c *ICalendar) WriteTo(filepath string) error {
	// Open or create a file at the specified path
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	// Initialises boilerplate
	lines := []string{
		"BEGIN:VCALENDAR\n",
		"VERSION:2.0\n",
		fmt.Sprintf("PRODID:-//%s/%s//NONSGML v1.0//EN\n", c.Coorperation, c.ProductName),
	}

	// Ranges over the events in the ICalendar, and creates ICal strings from them
	for _, event := range c.Events {
		eventString, err := event.ToString()
		if err != nil {
			return err
		}
		lines = append(lines, eventString)
	}

	// Writes the ICal strings to the file
	for _, line := range lines {
		f.WriteString(line)
	}

	// Ends Ical file with boilerplate
	f.WriteString("END:VCALENDAR")

	return nil
}
