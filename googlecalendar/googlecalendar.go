package googlecalendar

import (
	"context"
	"encoding/json"
	"fmt"
	"lectio-scraper/lectio"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type CalendarInfo struct {
	CalendarID string `json:"calendarID"`
}

type GoogleCalendar struct {
	Client       *http.Client
	Service      *calendar.Service
	CalendarInfo *CalendarInfo
}

func (googleCalendar *GoogleCalendar) Initialise(CalendarInfo *CalendarInfo) {
	ctx := context.Background()
	bytes, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Could not read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(bytes, calendar.CalendarEventsScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	googleCalendar.Client = GetClient(config)
	googleCalendar.CalendarInfo = CalendarInfo
	googleCalendar.Service, err = calendar.NewService(ctx, option.WithHTTPClient(googleCalendar.Client))
	if err != nil {
		log.Fatalf("Could not get Calendar client: %v", err)
	}
}

func GetClient(config *oauth2.Config) *http.Client {
	tokenFile := "token.json"
	token, err := tokenFromFile(tokenFile)
	if err != nil {
		token = getTokenFromWeb(config)
		saveToken(tokenFile, token)
	}
	return config.Client(context.Background(), token)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Printf("Go to the following link in your browser and type the authorization code %q\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Could not read authorization code: %v", err)
	}

	token, err := config.Exchange(context.TODO(), authCode)

	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	return token

}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func (googleCalendar *GoogleCalendar) AddModules(modules map[string]lectio.Module) {
	startTime := time.Now()
	defer log.Printf("Added modules to Google Calendar in %d", time.Since(startTime).Milliseconds())
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
		event, err := googleCalendar.Service.Events.Insert(googleCalendar.CalendarInfo.CalendarID, moduleEvent).Do()
		if err != nil {
			log.Fatalf("Unable to create event. %v\n", err)
		}
		fmt.Printf("Created event: %s\n", event.HtmlLink)
	}

}

func (googleCalendar *GoogleCalendar) GetModules(week int) map[string]lectio.Module {
	// events, err := googleCalendar.Service.Events.Get(googleCalendar.CalendarInfo.CalendarID)
	return make(map[string]lectio.Module)
}

func (GoogleCalendar *GoogleCalendar) CompareSchemes(l *lectio.Lectio) map[string]lectio.Module {

	// l.GetScheduleWeeks()

	return make(map[string]lectio.Module)
}
