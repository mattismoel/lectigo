package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

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

	var l *Lectio
	var c *GoogleCalendar

	// Creates command line flags and parses them upon execution of application
	command := flag.String("command", "sync", "what command should be executed")
	weekCount := flag.Int("weeks", 2, "define amount of weeks to sync Google Calendar")
	flag.Parse()

	// Checks for the provided commands and creates instances of Lectio and GoogleCalendar only if nescessary
	switch *command {
	// If user wants to sync Lectio schedule with Google Calendar
	// Creates appropriate clients, fetches modules and updates the Google Calendar
	case "sync":
		log.Printf("Syncing Google Calendar for the next %v weeks...\n", *weekCount)
		l = NewLectio(&lectioLoginInfo)                 // Creates a new Lectio client
		c = NewGoogleCalendar(&googleCalendarConfig)    // Creates a new Google Calendar client
		lectioModules := l.GetScheduleWeeks(*weekCount) // Gets the modules from the Lectio schedule
		googleModules := c.GetModules(*weekCount)       // Gets the modules present in Google Calendar
		c.UpdateCalendar(lectioModules, googleModules)  // Updates and deletes events that are missing in Google Calendar
	// If user wants to clear the Google Calendar
	// Create Google Calendar client and clear the calendar
	case "clear":
		log.Printf("Clearing Google Calendar...\n")
		c = NewGoogleCalendar(&googleCalendarConfig) // Creates a new Google Calendar client
		c.Clear()
	}
}
