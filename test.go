package lectigo

import (
	"log"
	"time"

	"github.com/mattismoel/icalendar"
)

func main () {	
	ical := icalendar.New("LectioSync", "LectioSync", "./icalendar.ics")
	event := &icalendar.ICalEvent{
		UID: "11223344@example.com",
		Summary: "Test",
		StartDate: time.Now(),
		EndDate: time.Now().Add(1 * time.Hour),
		Location: "06",
		Description: "Lorem ipsum dolor sit amet, qui minim labore adipisicing minim sint cillum sint consectetur cupidatat.",
	}
	
	for i := 0; i < 10; i++ {
		ical.Events = append(ical.Events, event)
	}

	err := ical.Update()
	if err != nil {
		log.Fatalf("could not write to icalendar.ics: %v\n", err)
	}
}
