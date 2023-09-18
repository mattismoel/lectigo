package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
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

func NewGoogleCalendar(CalendarInfo *CalendarInfo) *GoogleCalendar {
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

	return &GoogleCalendar{
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
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOnline)

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

func (c *GoogleCalendar) AddModules(modules map[string]Module) {
	startTime := time.Now()
	moduleCount := 0
	wg := sync.WaitGroup{}

	for key, module := range modules {
		wg.Add(1)
		go func(key string, module Module) {
			defer wg.Done()
			start := &calendar.EventDateTime{DateTime: module.StartDate.Format(time.RFC3339), TimeZone: "Europe/Copenhagen"}
			end := &calendar.EventDateTime{DateTime: module.EndDate.Format(time.RFC3339), TimeZone: "Europe/Copenhagen"}

			//Find color ID depending on the status of the module. "aflyst" results in red, "ændret" results in green
			calendarColorID := ""
			switch module.ModuleStatus {
			case "aflyst":
				calendarColorID = "4"
			case "ændret":
				calendarColorID = "2"
			}
			calEvent := &calendar.Event{
				Id:          "lec" + key,
				Start:       start,
				End:         end,
				ColorId:     calendarColorID,
				Summary:     module.Title,
				Description: module.Teacher,
				Location:    module.Room,
				Status:      "confirmed",
			}

			_, err := c.Service.Events.Update(googleCalendarConfig.CalendarID, calEvent.Id, calEvent).Do()
			if err != nil {
				log.Fatalf("Could not update event %v: %v\n", calEvent.Id, err)
			}
			moduleCount++

			// If event already exists in Google Calendar
			// if event, err := c.Service.Events.Get(googleCalendarConfig.CalendarID, calEvent.Id).Do(); err == nil {
			// 	if event.Status == "cancelled" {
			// 		log.Printf("Found deleted event %q\n", key)
			// 		_, err := c.Service.Events.Update(googleCalendarConfig.CalendarID, calEvent.Id, calEvent).Do()
			// 		if err != nil {
			// 			log.Fatalf("Could not update deleted event: %v\n", err)
			// 		}
			// 		moduleCount++
			// 		return
			// 	}
			// } else {
			// 	_, err := c.Service.Events.Insert(googleCalendarConfig.CalendarID, calEvent).Do()
			// 	if err != nil {
			// 		log.Fatalf("Could not insert event %q: %v\n", calEvent.Id, err)
			// 	}
			// 	moduleCount++
			// }
		}(key, module)
	}
	wg.Wait()

	// If no modules have been updated or inserted
	if !(moduleCount > 0) {
		log.Printf("Nothing to do. Lectio schedule is up to date with Google Calendar.\n")
		return
	}

	log.Printf("Added or updated %v modules to Google Calendar in %v\n", moduleCount, time.Since(startTime))

}

// func (c *GoogleCalendar) AddModules(modules map[string]Module) {
// 	startTime := time.Now()
// 	moduleCount := 0
// 	wg := sync.WaitGroup{}

// 	for key, module := range modules {
// 		wg.Add(1)
// 		go func(key string, module Module) {
// 			defer wg.Done()
// 			start := &calendar.EventDateTime{DateTime: module.StartDate.Format(time.RFC3339), TimeZone: "Europe/Copenhagen"}
// 			end := &calendar.EventDateTime{DateTime: module.EndDate.Format(time.RFC3339), TimeZone: "Europe/Copenhagen"}

// 			//Find color ID depending on the status of the module. "aflyst" results in red, "ændret" results in green
// 			calendarColorID := ""
// 			switch module.ModuleStatus {
// 			case "aflyst":
// 				calendarColorID = "4"
// 			case "ændret":
// 				calendarColorID = "2"
// 			}
// 			calEvent := &calendar.Event{
// 				Id:          "lec" + key,
// 				Start:       start,
// 				End:         end,
// 				ColorId:     calendarColorID,
// 				Summary:     module.Title,
// 				Description: module.Teacher,
// 				Location:    module.Room,
// 				Status:      "confirmed",
// 			}
// 			// If event already exists in Google Calendar
// 			if event, err := c.Service.Events.Get(googleCalendarConfig.CalendarID, calEvent.Id).Do(); err == nil {
// 				if event.Status == "cancelled" {
// 					log.Printf("Found deleted event %q\n", key)
// 					_, err := c.Service.Events.Update(googleCalendarConfig.CalendarID, calEvent.Id, calEvent).Do()
// 					if err != nil {
// 						log.Fatalf("Could not update deleted event: %v\n", err)
// 					}
// 					moduleCount++
// 					return
// 				}
// 			} else {
// 				_, err := c.Service.Events.Insert(googleCalendarConfig.CalendarID, calEvent).Do()
// 				if err != nil {
// 					log.Fatalf("Could not insert event %q: %v\n", calEvent.Id, err)
// 				}
// 				moduleCount++
// 			}
// 		}(key, module)
// 	}
// 	wg.Wait()

// 	// If no modules have been updated or inserted
// 	if !(moduleCount > 0) {
// 		log.Printf("Nothing to do. Lectio schedule is up to date with Google Calendar.\n")
// 		return
// 	}

// 	log.Printf("Added or updated %v modules to Google Calendar in %v\n", moduleCount, time.Since(startTime))

// }

// Returns all modules from Google Calendar.
func (c *GoogleCalendar) GetModules(weekCount int) map[string]Module {
	// startDate := RoundDateToDay(GetMonday())
	// endDate := RoundDateToDay(startDate.AddDate(0, 0, 7))

	googleCalModules := make(map[string]Module)

	s := time.Now()
	pageToken := ""
	eventCount := 0

	wg := sync.WaitGroup{}
	mu := sync.RWMutex{}

	startDate := GetMonday()
	endDate := GetMonday().AddDate(0, 0, 7*weekCount)
	req := c.Service.Events.List(c.CalendarInfo.CalendarID).TimeMin(startDate.Format(time.RFC3339)).TimeMax(endDate.Format(time.RFC3339))
	for {
		if pageToken != "" {
			req.PageToken(pageToken)
		}
		r, err := req.Do()
		if err != nil {
			log.Fatalf("Could not retrieve events: %v\n", err)
		}
		for _, item := range r.Items {
			if strings.Contains(item.Id, "lec") {
				wg.Add(1)
				go func(item *calendar.Event) {
					defer wg.Done()
					defer mu.Unlock()
					mu.Lock()

					id := strings.TrimPrefix(item.Id, "lec")
					googleCalModules[id] = GoogleEventToModule(item)
					eventCount++
				}(item)
			}
		}

		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}
	wg.Wait()
	log.Printf("Found %v events in %v\n", eventCount, time.Since(s))

	// =====

	// events, err := c.Service.Events.List(c.CalendarInfo.CalendarID).ShowDeleted(true).Do()
	// if err != nil {
	// 	log.Fatalf("Could not list the events of the calendar with ID %q: %v\n", c.CalendarInfo.CalendarID, err)
	// }
	// for _, event := range events.Items {
	// 	// If Google Calendar event is not a lectio module
	// 	if !strings.Contains(event.Id, "lec") {
	// 		continue
	// 	}
	// }

	// if err != nil {
	// 	log.Fatalf("Could not load location: %v\n", err)
	// }

	// for _, event := range events.Items {
	// 	startTime, err := time.Parse(time.RFC3339, event.Start.DateTime)
	// 	if err != nil {
	// 		// The event is an all-day event - skip
	// 		log.Printf("%v: Could not parse the date: %v\n", event.Summary, err)
	// 		continue
	// 	}

	// 	endTime, err := time.Parse(time.RFC3339, event.End.DateTime)
	// 	if err != nil {
	// 		log.Printf("%v: Could not parse the date: %v\n", event.Summary, err)
	// 	}

	// 	id := strings.TrimPrefix(event.Id, "lec")
	// 	googleCalModules[id] = Module{
	// 		Title:     event.Summary,
	// 		StartDate: startTime,
	// 		EndDate:   endTime,
	// 		Room:      event.Location,
	// 		Teacher:   event.Description,
	// 		Homework:  event.Description,
	// 	}
	// }
	return googleCalModules
}

func (c *GoogleCalendar) UpdateCalendar(lectioModules map[string]Module, googleModules map[string]Module) {
	// Finds the missing and extra modules in the Google Calendar with respect to the modules in the Lectio schedule
	extras, missing := CompareMaps(lectioModules, googleModules)

	for key := range extras {
		calID := "lec" + key
		err := c.Service.Events.Delete(c.CalendarInfo.CalendarID, calID).Do()
		if err != nil {
			log.Fatalf("Could not delete event %v: %v\n", calID, err)
		}
		log.Printf("Deleted %v\n", calID)
	}

	fmt.Println()
	fmt.Println("RESULTS", strings.Repeat("=", 23))
	fmt.Printf("%-30s%-10v\n", "Missing from Google Calendar:", len(missing))
	fmt.Printf("%-30s%-10v\n", "Extra in Google Calendar:", len(extras))
	fmt.Println(strings.Repeat("=", 31))
	fmt.Println()

	c.AddModules(missing)
}

func GoogleEventToModule(event *calendar.Event) Module {
	start, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		log.Fatalf("Could not parse start date: %v\n", err)
	}

	end, err := time.Parse(time.RFC3339, event.End.DateTime)
	if err != nil {
		log.Fatalf("Could not parse end date: %v\n", err)
	}
	return Module{
		Title:     event.Summary,
		StartDate: start,
		EndDate:   end,
		Room:      event.Location,
		Teacher:   event.Description,
		Homework:  event.Description,
	}
}

func (c *GoogleCalendar) Clear() {
	s := time.Now()
	pageToken := ""
	eventCount := 0

	wg := sync.WaitGroup{}

	for {
		req := c.Service.Events.List(c.CalendarInfo.CalendarID)
		if pageToken != "" {
			req.PageToken(pageToken)
		}
		r, err := req.Do()
		if err != nil {
			log.Fatalf("Could not retrieve events: %v\n", err)
		}
		for _, item := range r.Items {
			if strings.Contains(item.Id, "lec") {
				wg.Add(1)
				go func(item *calendar.Event) {
					defer wg.Done()
					err := c.Service.Events.Delete(c.CalendarInfo.CalendarID, item.Id).Do()
					if err != nil {
						log.Fatalf("Could not delete event %v: %v\n", item.Id, err)
					}
					eventCount++
				}(item)
			}
		}

		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}
	wg.Wait()
	log.Printf("Found and deleted %v events in %v\n", eventCount, time.Since(s))
}
