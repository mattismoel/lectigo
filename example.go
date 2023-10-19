package lectigo

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mattismoel/lectigo/util"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

func _main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Could not load environment variables: %v\n", err)
	}

	lectioPassword := os.Getenv("LECTIO_PASSWORD")
	lectioUsername := os.Getenv("LECTIO_USERNAME")
	lectioSchoolID := os.Getenv("LECTIO_SCHOOL_ID")

	googleCalendarID := os.Getenv("GOOGLE_CALENDAR_ID")

	// Reads the credentials file and creates a config from it - this is used to create the client
	bytes, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Could not read contents of credentials.json: %v\n", err)
	}

	config, err := google.ConfigFromJSON(bytes, calendar.CalendarEventsScope)
	if err != nil {
		log.Fatalf("Could not create config from credentials.json")
	}

	client, err := util.GetClient(config)
	if err != nil {
		log.Fatalf("Could not get Google Calendar client: %v\n", err)
	}

	c, err := NewGoogleCalendar(client, googleCalendarID)
	if err != nil {
		log.Fatalf("Could not create Google Calendar instance: %v\n", err)
	}
	l, err := NewLectio(&LectioLoginInfo{
		Username: lectioUsername,
		Password: lectioPassword,
		SchoolID: lectioSchoolID,
	})
	if err != nil {
		log.Fatalf("Could not create Lectio instance: %v\n", err)
	}

	lModules, err := l.GetScheduleWeeks(2)
	gEvents, err := c.GetEvents(2)
	err = c.UpdateCalendar(lModules, gEvents)
	if err != nil {
		log.Fatalf("Could not update Google Calendar: %v\n", err)
	}
}
