package lectigo

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mattismoel/lectigo/util"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Base struct for a Google Calendar client
type GoogleCalendar struct {
	Service *calendar.Service
	ID      string
	Logger  *log.Logger
}

// Base Google Calendar event struct.
type GoogleEvent calendar.Event

// Creates a new Google Calendar struct instance
func NewGoogleCalendar(client *http.Client, calendarID string) (*GoogleCalendar, error) {
	ctx := context.Background()

	service, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	calendar := &GoogleCalendar{
		Service: service,
		ID:      calendarID,
		Logger:  log.New(os.Stdout, "google-calendar ", log.LstdFlags),
	}
	return calendar, nil
}

// Returns all modules from Google Calendar.
func (c *GoogleCalendar) GetEvents(weekCount int) (map[string]*GoogleEvent, error) {
	googleCalModules := make(map[string]*GoogleEvent)
	pageToken := ""
	eventCount := 0

	wg := sync.WaitGroup{}
	mu := sync.RWMutex{}

	startDate, err := util.GetMonday()
	if err != nil {
		return nil, err
	}
	endDate := startDate.AddDate(0, 0, 7*weekCount)
	req := c.Service.Events.List(c.ID).ShowDeleted(true).TimeMin(startDate.Format(time.RFC3339)).TimeMax(endDate.Format(time.RFC3339))
	for {
		if pageToken != "" {
			req.PageToken(pageToken)
		}
		r, err := req.Do()
		if err != nil {
			return nil, err
		}
		for _, item := range r.Items {
			if strings.HasPrefix(item.Id, "lec") {
				wg.Add(1)
				go func(item *calendar.Event) {
					defer wg.Done()
					defer mu.Unlock()
					mu.Lock()
					gEvent := GoogleEvent(*item)
					googleCalModules[item.Id] = &gEvent
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
	return googleCalModules, nil
}

// Updates the Google Calendar with the input Lectio modules and Google Calendar events. The modules input should not be filtered, as the functions handles that (input all modules from Lectio and all events from Google Calendar)
func (c *GoogleCalendar) UpdateCalendar(lectioModules map[string]Module, googleEvents map[string]*GoogleEvent) error {
	var inserted int // For keeping track of inserted events count after execution
	var updated int  // For keeping track of updated events count after execution
	var deleted int  // For keeping track of deleted events count after execution

	startTime := time.Now()
	// Loops through each module in the Lectio schedule and checks for differences between it and the Google Calendar
	// If a Google Event is outdated, it is updated
	// If a Lectio module is missing from Google Calendar, it is inserted
	var wg sync.WaitGroup

	for lectioKey, lectioModule := range lectioModules {
		wg.Add(1)
		go func(lKey string, lModule Module) error {
			defer wg.Done()
			// If Lectio module is in Google Calendar
			key := "lec" + lKey
			if _, ok := googleEvents[key]; ok {
				googleEvent := *googleEvents[key]
				googleModule, err := googleEvent.ToModule()
				if err != nil {
					return err
				}
				needsUpdate := !lModule.Equals(googleModule)
				isCancelled := googleEvent.Status == "cancelled"

				if needsUpdate || isCancelled {
					c.Logger.Printf("Attempting to update %v\n", googleEvent.Id)
					lectioEvent := calendar.Event(*lModule.ToGoogleEvent())
					_, err := c.Service.Events.Update(c.ID, googleEvent.Id, &lectioEvent).Do()
					if err != nil {
						return err
					}
					updated++
				} else {
					return nil
				}
			} else {
				googleEvent := calendar.Event(*lModule.ToGoogleEvent())
				_, err := c.Service.Events.Insert(c.ID, &googleEvent).Do()
				if err != nil {
					return err
				}
				inserted++
			}
			return nil
		}(lectioKey, lectioModule)
	}

	wg.Wait()

	// Loops through all Google Events and checks if it should be deleted
	for googleKey, googleEvent := range googleEvents {
		wg.Add(1)
		go func(googleKey string, googleEvent *GoogleEvent) error {
			defer wg.Done()
			trimPrefix := strings.TrimPrefix(googleKey, "lec")

			if _, ok := lectioModules[trimPrefix]; !ok {
				if googleEvent.Status != "cancelled" {
					c.Logger.Printf("Attempting to delete %v\n", googleKey)
					err := c.Service.Events.Delete(c.ID, googleKey).Do()
					if err != nil {
						return err
					}
					deleted++
				}
			}
			return nil
		}(googleKey, googleEvent)
	}
	wg.Wait()

	fmt.Printf(`
RESULTS ==============================
UPDATED %v events in Google Calendar
INSERTED %v events into Google Calendar
DELETED %v events from Google Calendar

Execution took %v
======================================`,
		updated, inserted, deleted, time.Since(startTime))
	return nil
}

// Clears the Google Calendar of Lectigo events
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
				go func(item *calendar.Event) error {
					defer wg.Done()
					err := c.Service.Events.Delete(c.ID, item.Id).Do()
					if err != nil {
						return err
						// log.Fatalf("Could not delete event %v: %v\n", item.Id, err)
					}
					eventCount++
					return nil
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

// Converts a Google Calendar event to a Lectio module
func (e *GoogleEvent) ToModule() (*Module, error) {
	location, err := time.LoadLocation("Europe/Copenhagen")
	if err != nil {
		return nil, err
	}
	start, err := time.ParseInLocation(time.RFC3339, e.Start.DateTime, location)
	if err != nil {
		return nil, err
	}

	end, err := time.ParseInLocation(time.RFC3339, e.End.DateTime, location)
	if err != nil {
		return nil, err
	}

	var homework, teacher string
	
	re := regexp.MustCompile(`Lærer: \[(.*?)\]\nLektier:\n\[(.*?)\]`)
	matches := re.FindStringSubmatch(e.Description)

	if len(matches) == 3 {
		teacher = matches[1]
		homework = matches[2]
	}

	module := &Module{
		Id:           strings.TrimPrefix(e.Id, "lec"),
		Title:        e.Summary,
		StartDate:    start,
		EndDate:      end,
		Room:         e.Location,
		Teacher:      teacher,
		Homework:     homework,
		ModuleStatus: util.StatusFromColorID(e.ColorId),
	}

	return module, nil
}
