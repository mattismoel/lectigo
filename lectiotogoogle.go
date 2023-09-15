package main

import (
	"encoding/json"
	"lectio-scraper/googlecalendar"
	"lectio-scraper/lectio"
	"log"
	"os"
)

type SecretsConfig struct {
	UserInfo                    lectio.LectioLoginInfo
	GoogleCalendarConfiguration googlecalendar.CalendarInfo
}

// The login information of the user. This should be stored in a "secrets.json" file, and should have the following variables: username, password, schoolID, calendarID
var lectioLoginInfo lectio.LectioLoginInfo
var googleCalendarConfig googlecalendar.CalendarInfo

func main() {

	// Reads the content of the lectioSecrets.json file and attempts to unmarshal it to the lectioLoginInfo variable.
	// This stores the users login information
	b, err := os.ReadFile("lectioSecrets.json")
	if err != nil {
		log.Fatalf("Could not read the contents of %q: %v\n", "lectioSecrets.json", err)
	}
	if err := json.Unmarshal(b, &lectioLoginInfo); err != nil {
		panic(err)
	}

	// Reads the content of the lectioSecrets.json file and attempts to unmarshal it to the lectioLoginInfo variable.
	// This stores the users login information
	b, err = os.ReadFile("googleSecrets.json")
	if err != nil {
		log.Fatalf("Could not read the contents of %q: %v\n", "googleSecrets.json", err)
	}
	if err := json.Unmarshal(b, &googleCalendarConfig); err != nil {
		panic(err)
	}

	lectio := lectio.Lectio{}
	lectio.Initialise(&lectioLoginInfo)
	googleCalendar := googlecalendar.GoogleCalendar{}
	googleCalendar.Initialise(&googleCalendarConfig)

	modules := lectio.GetScheduleWeeks(1, true)
	googleCalendar.AddModules(modules)
}
