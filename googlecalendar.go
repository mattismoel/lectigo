package main

import (
	"context"
	"encoding/json"
	"fmt"
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

func NewGoogleCalendar(CalendarInfo *CalendarInfo) GoogleCalendar {
	ctx := context.Background()
	bytes, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Could not read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(bytes, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := *GetClient(config)
	service, err := calendar.NewService(ctx, option.WithHTTPClient(&client))
	if err != nil {
		log.Fatalf("Could not get Calendar client: %v", err)
	}

	return GoogleCalendar{
		Client:       &client,
		Service:      service,
		CalendarInfo: CalendarInfo,
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

func (googleCalendar *GoogleCalendar) AddModules(lectioModules map[string]Module, googleModules map[string]Module) {
	startTime := time.Now()
	defer log.Printf("Added modules to Google Calendar in %d", time.Since(startTime).Milliseconds())

	for key, module := range lectioModules {
		start := &calendar.EventDateTime{DateTime: module.StartDate.Format(time.RFC3339), TimeZone: "Europe/Copenhagen"}
		end := &calendar.EventDateTime{DateTime: module.EndDate.Format(time.RFC3339), TimeZone: "Europe/Copenhagen"}

		// Get the correct color ID for the Google Calendar event - red for "aflyst", green for "ændret"
		calendarColorID := ""
		switch module.Status {
		case "aflyst":
			calendarColorID = "4"
		case "ændret":
			calendarColorID = "2"
		}

		// If event already exists
		if _, err := googleCalendar.Service.Events.Get(googleCalendar.CalendarInfo.CalendarID, key).Do(); err == nil {
			_, updateErr := googleCalendar.Service.Events.Update(googleCalendar.CalendarInfo.CalendarID, key, &calendar.Event{
				Start:       start,
				End:         end,
				Summary:     module.Title,
				ColorId:     calendarColorID,
				Location:    module.Room,
				Description: module.Teacher,
			}).Do()

			if updateErr != nil {
				log.Fatalf("Could not update module: %v\n", updateErr)
			}
			log.Printf("Updated event %v", key)
		} else {
			_, insertErr := googleCalendar.Service.Events.Insert(googleCalendar.CalendarInfo.CalendarID, &calendar.Event{
				Id:          key,
				ColorId:     calendarColorID,
				Start:       start,
				End:         end,
				Summary:     module.Title,
				Location:    module.Room,
				Description: module.Teacher,
			}).Do()
			if insertErr != nil {
				log.Fatalf("Could not insert event %v: %v\n", key, insertErr)
			}
		}
	}
}

func (googleCalendar *GoogleCalendar) GetModules(weekCount int) map[string]Module {
	startDate := RoundDateToDay(GetMonday())
	endDate := RoundDateToDay(startDate.AddDate(0, 0, 7))
	events, err := googleCalendar.Service.Events.List(googleCalendar.CalendarInfo.CalendarID).TimeMin(startDate.Format(time.RFC3339)).TimeMax(endDate.Format(time.RFC3339)).ShowDeleted(true).Do()
	if err != nil {
		log.Fatalf("Could not list the events of the calendar with ID %q: %v\n", googleCalendar.CalendarInfo.CalendarID, err)
	}

	googleCalModules := make(map[string]Module)

	if err != nil {
		log.Fatalf("Could not load location: %v\n", err)
	}

	// _, currWeek := time.Now().ISOWeek()
	for _, event := range events.Items {
		startTime, err := time.Parse(time.RFC3339, event.Start.DateTime)
		if err != nil {
			// The event is an all-day event - skip
			log.Printf("%v: Could not parse the date: %v\n", event.Summary, err)
			continue
		}

		endTime, err := time.Parse(time.RFC3339, event.End.DateTime)
		if err != nil {
			log.Printf("%v: Could not parse the date: %v\n", event.Summary, err)
		}

		googleCalModules[event.Id] = Module{
			Id:        event.Id,
			Title:     event.Summary,
			StartDate: startTime,
			EndDate:   endTime,
			Room:      event.Location,
			Teacher:   event.Description,
			Homework:  event.Description,
		}
	}

	// fmt.Println(googleCalModules)
	return googleCalModules
}

func (googleCalendar *GoogleCalendar) UpdateCalendar(lectioModules map[string]Module, googleModules map[string]Module) {
	// Finds the missing and extra modules in the Google Calendar with respect to the modules in the Lectio schedule
	extras, missing := CompareMaps(lectioModules, googleModules)

	for _, miss := range missing {
		fmt.Println(miss.Id)
	}
	// Delets all the extra events from the Google Calendar
	for id := range extras {
		fmt.Println("EXTRA")
		if err := googleCalendar.Service.Events.Delete(googleCalendar.CalendarInfo.CalendarID, id).Do(); err != nil {
			log.Fatalf("Could not delete extra event: %v\n", err)
		}
		log.Printf("Deleted removed event %q\n", id)
	}

	googleCalendar.AddModules(lectioModules, googleModules)
	fmt.Println("LENGTH:", len(missing))
}

func GoogleEventToModule(event *calendar.Event) Module {
	start, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		log.Fatal("Could not parse start date: %v\n", err)
	}

	end, err := time.Parse(time.RFC3339, event.End.DateTime)
	if err != nil {
		log.Fatal("Could not parse end date: %v\n", err)
	}
	return Module{
		Id:        event.Id,
		Title:     event.Summary,
		StartDate: start,
		EndDate:   end,
		Room:      event.Location,
		Teacher:   event.Description,
		Homework:  event.Description,
	}
}
