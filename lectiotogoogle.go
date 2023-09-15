package main

import (
	"encoding/json"
	"fmt"
	"lectio-scraper/googlecalendar"
	"lectio-scraper/lectio"
	"log"
	"os"
	"time"

	"google.golang.org/api/calendar/v3"
)

type SecretsConfig struct {
	UserInfo                    lectio.LectioLoginInfo
	GoogleCalendarConfiguration GoogleCalendarConfig
}

type GoogleCalendarConfig struct {
	CalendarID string `json:"calendarID"`
}

// The login information of the user. This should be stored in a "secrets.json" file, and should have the following variables: username, password, schoolID, calendarID
var lectioLoginInfo lectio.LectioLoginInfo
var googleCalendarConfig GoogleCalendarConfig

func main() {

	// Reads the content of the lectioSecrets.json file and attempts to unmarshal it to the lectioLoginInfo variable.
	// This stores the users login information
	b, err := os.ReadFile("lectioSecrets.json")
	if err != nil {
		log.Fatalf("Could not read the contents of %q: %v\n", "lectioSecrets.json", err)
	}
	if err := json.Unmarshal(b, &lectioLoginInfo); err != nil {
		fmt.Println("SLKDJFSDFLJK")
		// panic(err)
	}

	// Reads the content of the lectioSecrets.json file and attempts to unmarshal it to the lectioLoginInfo variable.
	// This stores the users login information
	b, err = os.ReadFile("googleSecrets.json")
	if err != nil {
		log.Fatalf("Could not read the contents of %q: %v\n", "googleSecrets.json", err)
	}
	if err := json.Unmarshal(b, &googleCalendarConfig); err != nil {
		fmt.Println("ELFDSLKJF")
		// panic(err)
	}

	lectio := lectio.Lectio{}
	lectio.Initialise(&lectioLoginInfo)
	googleCalendar := googlecalendar.GoogleCalendar{}
	googleCalendar.Initialise()

	modules := lectio.GetScheduleWeeks(1, false)
	AddToGoogleCalendar(&googleCalendar, modules, googleCalendarConfig.CalendarID)
}

func AddToGoogleCalendar(googleCalendar *googlecalendar.GoogleCalendar, modules []lectio.Module, calendarID string) {
	for _, module := range modules {
		start := &calendar.EventDateTime{DateTime: module.StartDate.Format(time.RFC3339), TimeZone: "Europe/Copenhagen"}
		end := &calendar.EventDateTime{DateTime: module.EndDate.Format(time.RFC3339), TimeZone: "Europe/Copenhagen"}

		calendarColorID := ""
		switch module.Status {
		case "aflyst":
			calendarColorID = "4"
		case "ændret":
			calendarColorID = "2"

		}
		moduleEvent := &calendar.Event{
			Start:       start,
			End:         end,
			Summary:     module.Title,
			Location:    module.Room,
			Description: fmt.Sprintf("Lærer: %s\n%s\n", module.Teacher, module.Homework),
			ColorId:     calendarColorID,
		}
		_, err := time.Parse(time.RFC3339, moduleEvent.Start.DateTime)
		if err != nil {
			log.Fatalf("Could not parse date: %v\n", err)
		}
		event, err := googleCalendar.Service.Events.Insert(calendarID, moduleEvent).Do()
		if err != nil {
			log.Fatalf("Unable to create event. %v\n", err)
		}
		fmt.Printf("Created event: %s\n", event.HtmlLink)
	}

	fmt.Println("Added modules to Calendar")
}
