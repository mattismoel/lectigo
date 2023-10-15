package types

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/mattismoel/lectigo/util"
	"google.golang.org/api/calendar/v3"
)

type GoogleCalendar struct {
	Service *calendar.Service
	ID      string
	Logger  *log.Logger
}

type GoogleEvent struct {
	event *calendar.Event
}

// Returns all modules from Google Calendar.
func (c *GoogleCalendar) GetModules(weekCount int) (map[string]*calendar.Event, error) {
	googleCalModules := make(map[string]*calendar.Event)
	s := time.Now()
	pageToken := ""
	eventCount := 0

	wg := sync.WaitGroup{}
	mu := sync.RWMutex{}

	startDate, err := util.GetMonday()
	if err != nil {
		return nil, err
	}
	endDate := startDate.AddDate(0, 0, 7*weekCount).Add(-24 * time.Hour)
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
					googleCalModules[item.Id] = item
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
	for key := range googleCalModules {
		_, err := c.Service.Events.Patch(c.ID, key, &calendar.Event{
			Start: &calendar.EventDateTime{
				DateTime: time.Now().Format(time.RFC3339),
				TimeZone: "Europe/Copenhagen",
			},

			End: &calendar.EventDateTime{
				DateTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				TimeZone: "Europe/Copenhagen",
			},
		}).Do()
		if err != nil {
			log.Fatalf("Could not update event with id %v: %v\n", key, err)
		}
	}
	return googleCalModules, nil
}

func (c *GoogleCalendar) UpdateCalendar(lectioModules map[string]Module, googleEvents map[string]*calendar.Event) error {
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
			gEvent := lModule.ToGoogleEvent()
			key := "lec" + lKey
			// fmt.Printf("_____\nid: %v\nstatus: %v\n_____\n", key, gEvent.Status)
			if _, ok := googleEvents[key]; ok {
				startDate, err := time.Parse(time.RFC3339, gEvent.event.Start.DateTime)
				if err != nil {
					return err
				}

				monday, err := util.GetMonday()

				outOfBounds := startDate.Before(monday)

				// fmt.Printf("%v", PrettyPrint(googleEvent))
				// fmt.Println("in calendar", gEvent.Id, gEvent.Start.DateTime)
				isCancelled := gEvent.event.Status == "cancelled"
				needsUpdate := gEvent.ToModule() != lModule

				if isCancelled {
					fmt.Printf("Cancelled: %v\n", gEvent.event.Id)
				}
				if needsUpdate {
					fmt.Printf("Needs update: %v\n", gEvent.event.Id)
				}
				if (isCancelled || needsUpdate) && !outOfBounds {
					c.Logger.Printf("Attempting to update %v\n", gEvent.event.Id)
					_, err := c.Service.Events.Update(c.ID, gEvent.event.Id, gEvent.event).Do()
					if err != nil {
						return err
					}
					updated++
					return nil
				}
				return nil
			} else {
				// event := lectioModuleToGoogleEvent(&lModule)
				// fmt.Printf("status: %v\n_____\n", gEvent.Status)
				// fmt.Println("not in calendar", key, gEvent.Start.DateTime)
				_, err := c.Service.Events.Insert(c.ID, gEvent.event).Do()
				if err != nil {
					log.Fatalf("DID NOT INSERT EVENT %v: %v\n", gEvent.event.Id, err)
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
		go func(googleKey string, googleEvent *calendar.Event) error {
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

	// Print statements for displaying results.
	// Prints missing and extra modules in the Google Calendar
	fmt.Println()
	fmt.Println("RESULTS", strings.Repeat("=", 23))
	fmt.Printf("Updated %v events in Google Calendar\n", updated)
	fmt.Printf("Deleted %v events from Google Calendar\n", deleted)
	fmt.Printf("Inserted %v events into Google Calendar\n", inserted)
	fmt.Println(strings.Repeat("=", 31))
	fmt.Printf("\nThe execution took %v\n\n", time.Since(startTime))

	return nil
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

func (c *GoogleCalendar) AddModules(modules map[string]Module) {
	startTime := time.Now()
	var updateCount, insertCount int
	wg := sync.WaitGroup{}

	for key, module := range modules {
		wg.Add(1)
		go func(key string, module Module) {
			defer wg.Done()
			calEvent := module.ToGoogleEvent()
			_, err := c.Service.Events.Insert(c.ID, calEvent.event).Do()
			if err != nil {
				c.Logger.Fatalf("Could not insert missing event: %v\n", err)
			}
			insertCount++
			c.Logger.Printf("Inserted new event\n")
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

func (e *GoogleEvent) ToModule() Module {
	start, err := time.Parse(time.RFC3339, e.event.Start.DateTime)
	if err != nil {
		log.Fatalf("Could not parse start date: %v\n", err)
	}

	end, err := time.Parse(time.RFC3339, e.event.End.DateTime)
	if err != nil {
		log.Fatalf("Could not parse end date: %v\n", err)
	}

	homework := ""
	// fmt.Println(event.Description)

	return Module{
		Id:           strings.TrimPrefix(e.event.Id, "lec"),
		Title:        e.event.Summary,
		StartDate:    start,
		EndDate:      end,
		Room:         e.event.Location,
		Teacher:      e.event.Description,
		Homework:     homework,
		ModuleStatus: util.StatusFromColorID(e.event.ColorId),
	}
}

// func (e *GoogleEvent) ToModule(event *calendar.Event) (module *Module, err error) {
// 	var teacher, homework, status string
// 	start, err := time.Parse(time.RFC3339, event.Start.DateTime)
// 	if err != nil {
// 		log.Fatalf("Could not parse date: %v\n", err)
// 	}
//
// 	end, err := time.Parse(time.RFC3339, event.End.DateTime)
// 	if err != nil {
// 		return nil, err
// 		// log.Fatalf("Could not parse end date: %v\n", err)
// 	}
// 	module = &Module{
// 		Id:           event.Id,
// 		Title:        event.Summary,
// 		Room:         event.Location,
// 		StartDate:    start,
// 		EndDate:      end,
// 		Teacher:      teacher,
// 		Homework:     homework,
// 		ModuleStatus: status,
// 	}
// 	return module, nil
// }
