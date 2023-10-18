package main

import (
	"log"
	"time"

	"github.com/mattismoel/lectigo/types"
)



func main () {	
	ical := &types.ICalendar{
		Coorperation: "LectioSync",
		ProductName: "LectioSync",
	}

	event := &types.ICalEvent{
		UID: "11223344@example.com",
		Summary: "Test",
		StartDate: time.Now(),
		EndDate: time.Now().Add(1 * time.Hour),
		Location: &types.Location{
			Lon: 55,
			Lat: 44,
		},
	}
	
	for i := 0; i < 10; i++ {
		ical.Events = append(ical.Events, event)
	}

	err := ical.WriteTo("./icalendar.ics")
	if err != nil {
		log.Fatalf("could not write to icalendar.ics: %v\n", err)
	}

	// f, err := icalendar.CreateICalendar("./icalendar.ics")
	// if err != nil {
	// 	log.Fatalf("could not write icalendar file: %v\n", err)
	// }
	// log.Println(f)
	// t := "19970715T040000Z"
	// date, err := util.ICalTimestampToTime(t)
	// if err != nil {
	// 	log.Fatalf("could not parse ical timestamp: %v\n", err)
	// }
	//
	// str, err := util.TimeToICalTimestamp(date)
	// if err != nil {
	// 	log.Fatalf("could not parse date to timestamp: %v\n", err)
	// }
	
	// fmt.Println(date)
	// fmt.Println(str)
	// fmt.Println(t)

}
