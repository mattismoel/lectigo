package main

import (
	"encoding/json"
	"log"
	"os"
)

type SecretsConfig struct {
	UserInfo                    LectioLoginInfo
	GoogleCalendarConfiguration CalendarInfo
}

// The login information of the user. This should be stored in a "secrets.json" file, and should have the following variables: username, password, schoolID, calendarID
var lectioLoginInfo LectioLoginInfo
var googleCalendarConfig CalendarInfo

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

	l := NewLectio(&lectioLoginInfo)
	googleCalendar := NewGoogleCalendar(&googleCalendarConfig)
	lectioModules := l.GetScheduleWeeks(2)

	// googleCalendar.Clear()
	googleCalendar.AddModules(lectioModules)
}
