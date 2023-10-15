package lectigo

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// The login information of the user. This should be stored in a "secrets.json" file, and should have the following variables: username, password, schoolID, calendarID
func main() {
	bytes, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Could not read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(bytes, calendar.CalendarScope)
	config.RedirectURL = "http://localhost:3000/oauth/token"

	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := *GetClient(config)

	lectioLoginInfo := &LectioLoginInfo{
		Username: "username",
		Password: "password",
		SchoolID: "schoolID",
	}

	l := NewLectio(lectioLoginInfo)
	c := NewGoogleCalendar(&client, "googleCalendarID")

	fmt.Println(l, c)
}
