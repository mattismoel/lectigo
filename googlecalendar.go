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

	"github.com/google/go-cmp/cmp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type GoogleCalendar struct {
	Service *calendar.Service
	ID      string
	l       *log.Logger
}

func NewGoogleCalendar(id string) *GoogleCalendar {
	ctx := context.Background()

	bytes, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Could not read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(bytes, calendar.CalendarScope)

	config.RedirectURL = "http://localhost:3000/oauth/token"

	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := *getClient(config)
	service, err := calendar.NewService(ctx, option.WithHTTPClient(&client))
	if err != nil {
		log.Fatalf("Could not get Calendar client: %v", err)
	}

	return &GoogleCalendar{
		Service: service,
		ID:      id,
		l:       log.New(os.Stdout, "google-calendar", log.LstdFlags),
	}
}

func getClient(config *oauth2.Config) *http.Client {
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
		log.Fatalf("Could not read authorization code: %v\n", err)
	}

	token, err := config.Exchange(context.TODO(), authCode)

	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v\n", err)
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

func (c *GoogleCalendar) addModules(modules map[string]Module) {
	startTime := time.Now()
	var updateCount, insertCount int

	wg := sync.WaitGroup{}
	// for _, module := range modules {
	// 	fmt.Printf(PrettyPrint(module))
	// }
	for key, module := range modules {
		wg.Add(1)
		go func(key string, module Module) {
			defer wg.Done()
			// start := &calendar.EventDateTime{DateTime: module.StartDate.Format(time.RFC3339), TimeZone: "Europe/Copenhagen"}
			// end := &calendar.EventDateTime{DateTime: module.EndDate.Format(time.RFC3339), TimeZone: "Europe/Copenhagen"}

			//Find color ID depending on the status of the module. "aflyst" results in red, "ændret" results in green

			calEvent := lectioModuleToGoogleEvent(&module)
			// calEvent := &calendar.Event{
			// 	Id:          "lec" + key,
			// 	Start:       start,
			// 	End:         end,
			// 	ColorId:     calendarColorID,
			// 	Summary:     module.Title,
			// 	Description: module.Teacher,
			// 	Location:    module.Room,
			// 	Status:      "confirmed",
			// }
			_, err := c.Service.Events.Insert(c.ID, calEvent).Do()
			if err != nil {
				c.l.Fatalf("Could not insert missing event: %v\n", err)
			}
			insertCount++
			c.l.Printf("Inserted new event\n")
		}(key, module)
	}
	wg.Wait()

	// If no modules have been updated or inserted
	if insertCount == 0 && updateCount == 0 {
		log.Printf("Nothing to do. Lectio schedule is up to date with Google Calendar.\n")
		return
	}

	log.Printf("Added %v modules and updated %v modules in Google Calendar in %v\n", insertCount, updateCount, time.Since(startTime))
}

// Returns all modules from Google Calendar.
func (c *GoogleCalendar) GetModules(weekCount int) (googleCalModules map[string]Module, err error) {
	googleCalModules = make(map[string]Module)

	s := time.Now()
	pageToken := ""
	eventCount := 0

	wg := sync.WaitGroup{}
	mu := sync.RWMutex{}

	startDate, err := GetMonday()
	if err != nil {
		return nil, err
	}
	endDate := startDate.AddDate(0, 0, 7*weekCount)
	fmt.Printf("MONDAY %v", startDate)
	req := c.Service.Events.List(c.ID).TimeMin(startDate.Format(time.RFC3339)).TimeMax(endDate.Format(time.RFC3339))
	for {
		if pageToken != "" {
			req.PageToken(pageToken)
		}
		r, err := req.Do()
		if err != nil {
			return nil, err
		}
		for _, item := range r.Items {
			if strings.Contains(item.Id, "lec") {
				wg.Add(1)
				go func(item *calendar.Event) {
					defer wg.Done()
					defer mu.Unlock()
					mu.Lock()

					id := strings.TrimPrefix(item.Id, "lec")
					googleCalModules[id] = googleEventToModule(item)
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
	return googleCalModules, nil
}

func (c *GoogleCalendar) UpdateCalendar(lectioModules map[string]Module, googleModules map[string]Module) error {
	// Finds the missing and extra modules in the Google Calendar with respect to the modules in the Lectio schedule
	extras, _ := CompareMaps(lectioModules, googleModules)

	// for _, miss := range missing {
	// 	fmt.Println(PrettyPrint(miss))
	// }
	// Deletes every module that is not in the Lectio schedule
	for key := range extras {
		calID := "lec" + key
		err := c.Service.Events.Delete(c.ID, calID).Do()
		if err != nil {
			return err
			// log.Fatalf("Could not delete event %v: %v\n", calID, err)
		}
		log.Printf("Deleted %v\n", calID)
	}

	missing := make(map[string]Module)

	// Print statements for displaying results.
	// Prints missing and extra modules in the Google Calendar
	fmt.Println()
	fmt.Println("RESULTS", strings.Repeat("=", 23))
	fmt.Printf("%-30s%-10v\n", "Missing from Google Calendar:", len(missing))
	fmt.Printf("%-30s%-10v\n", "Extra in Google Calendar:", len(extras))
	fmt.Println(strings.Repeat("=", 31))
	fmt.Println()

	for key, lectioModule := range lectioModules {
		// If lectio module is in Google Calendar
		if googleModule, ok := googleModules[key]; ok {
			// If the events are not the same - outdated, update the module
			if !cmp.Equal(googleModule, lectioModule) {
				fmt.Printf("===========\nShould be updated: %v\n\n\n VS: %v\n\n =============:", PrettyPrint(lectioModule), PrettyPrint(googleModule))
				if _, err := c.Service.Events.Update(c.ID, "lec"+key, lectioModuleToGoogleEvent(&lectioModule)).Do(); err != nil {
					return err
				}
			}
			continue
		}
		missing[key] = lectioModule
		// If not in Google Calendar, insert it
		// if _, err := c.Service.Events.Insert(c.ID, lectioModuleToGoogleEvent(&lectioModule)).Do(); err != nil {
		// 	return err
		// }
	}

	// Adds the missing modules to Google Calendar
	c.addModules(missing)
	return nil
}

func lectioModuleToGoogleEvent(m *Module) *calendar.Event {
	calendarColorID := ""
	switch m.ModuleStatus {
	case "aflyst":
		calendarColorID = "4"
	case "ændret":
		calendarColorID = "2"
	}
	return &calendar.Event{
		Id:          "lec" + m.Id,
		Description: m.Teacher,
		Start: &calendar.EventDateTime{
			DateTime: m.StartDate.Format(time.RFC3339),
			TimeZone: "Europe/Copenhagen",
		},
		End: &calendar.EventDateTime{
			DateTime: m.EndDate.Format(time.RFC3339),
			TimeZone: "Europe/Copenhagen",
		},
		Location: m.Room,
		Summary:  m.Title,
		ColorId:  calendarColorID,
	}
}

func googleEventToModule(event *calendar.Event) Module {
	start, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		log.Fatalf("Could not parse start date: %v\n", err)
	}

	end, err := time.Parse(time.RFC3339, event.End.DateTime)
	if err != nil {
		log.Fatalf("Could not parse end date: %v\n", err)
	}
	return Module{
		Id:           strings.TrimPrefix(event.Id, "lec"),
		Title:        event.Summary,
		StartDate:    start,
		EndDate:      end,
		Room:         event.Location,
		Teacher:      event.Description,
		Homework:     "homework",
		ModuleStatus: statusFromColorID(event.ColorId),
	}
}

func (c *GoogleCalendar) Clear() error {
	s := time.Now()
	pageToken := ""
	eventCount := 0

	wg := sync.WaitGroup{}

	for {
		req := c.Service.Events.List(c.ID)
		if pageToken != "" {
			req.PageToken(pageToken)
		}
		r, err := req.Do()
		if err != nil {
			return err
			// log.Fatalf("Could not retrieve events: %v\n", err)
		}
		for _, item := range r.Items {
			if strings.Contains(item.Id, "lec") {
				wg.Add(1)
				go func(item *calendar.Event) {
					defer wg.Done()
					err := c.Service.Events.Delete(c.ID, item.Id).Do()
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
	return nil
}

func statusFromColorID(colorId string) string {
	switch colorId {
	case "4":
		return "aflyst"
	case "2":
		return "ændret"
	}
	return "uændret"
}
