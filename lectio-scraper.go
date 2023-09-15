package main

import (
	"context"
	"encoding/json"
	"fmt"
	"lectio-scraper/googlecalendar"
	"lectio-scraper/utils"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Module struct {
	Title     string    `json:"title"`     // Title of the module (eg. 3a Dansk)
	StartDate time.Time `json:"startDate"` // The start date of the module. This includes the date as well as the time of start (eg. 09:55)
	EndDate   time.Time `json:"endDate"`   // The end date of the module. This includes the date as well as the time of end (eg. 11:25)
	Room      string    `json:"room"`      // The room of the module (eg. 22)
	Teacher   string    `json:"teacher"`   // The teacher of the class
	Homework  string    `json:"homework"`  // Homework for the module
	Status    string    `json:"status"`    // The status of the module (eg. "Ændret" or "Aflyst")
}

var userName string = "" // Username of user
var password string = "" // Password of user
var schoolID string = "" // School ID of user. This can be found on the logged on homepage of Lectio (eg. www.lectio.dk/lectio/<id>/SkemaNy.aspx)

func main() {
	startTime := time.Now()
	c := colly.NewCollector(colly.AllowedDomains("lectio.dk", "www.lectio.dk"))
	loginUrl := fmt.Sprintf("https://www.lectio.dk/lectio/%s/login.aspx", schoolID)
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	authToken := utils.GetToken(loginUrl, client)
	// Attempts to log the user in with the given login information
	err := c.Post(loginUrl, map[string]string{
		"m$Content$username": userName,
		"m$Content$password": password,
		"__EVENTVALIDATION":  authToken.Token,
		"__EVENTTARGET":      "m$Content$submitbtn2",
		"__EVENTARGUMENT":    "",
		"masterfootervalue":  "X1!ÆØÅ",
		"LectioPostbackId":   "",
	})

	if err != nil {
		log.Fatal("Could not log the user in. Please check that the login information is correct", err)
	}

	// Function is fired when a new page is loaded
	c.OnResponse(func(r *colly.Response) {
		log.Println("response received", r.StatusCode, r.Request.URL)
	})

	err = c.Visit("https://www.lectio.dk/lectio/143/forside.aspx")
	if err != nil {
		log.Fatalf("Could not visit %s. %s", "https://www.lectio.dk/lectio/143/forside.aspx", err)
	}
	modules := getScheduleWeeks(c, 2, false)
	AddToGoogleCalendar(modules, "")
	for _, module := range getScheduleWeeks(c, 1, true) {
		fmt.Println(module)
	}
	fmt.Printf("Ran in %v", time.Since(startTime))
}

func getScheduleWeeks(c *colly.Collector, weekCount int, toJSON bool) []Module {
	modules := []Module{}

	for i := 0; i < weekCount; i++ {
		_, week := time.Now().ISOWeek()
		weekModules := getSchedule(c, week+i)
		modules = append(modules, weekModules...)
	}

	if toJSON && len(modules) > 0 {
		b, err := json.Marshal(modules)
		if err != nil {
			log.Fatal("Could not marshal JSON.", err)
		}

		err = os.WriteFile("schedule.json", b, 0644)
		if err != nil {
			log.Fatal("Could not write to file", err)
		}
	}
	return modules
}

// Returns the modules of a given week. Preferably the getScheduleWeeks() function is used.
func getSchedule(c *colly.Collector, week int) []Module {
	wg := sync.WaitGroup{}
	modules := []Module{}
	c.OnHTML("a.s2skemabrik.s2brik", func(e *colly.HTMLElement) {
		wg.Add(1)
		defer wg.Done()
		addInfo := e.Attr("data-additionalinfo")

		lines := strings.Split(addInfo, "\n")

		var title, teacher, room, homework string
		var status = "uændret"
		var startDate, endDate time.Time
		location, err := time.LoadLocation("Europe/Copenhagen")
		if err != nil {
			log.Fatalf("Could not load location: %s\n", err)
		}

		if strings.Contains(addInfo, "Lektier:") {
			_, homework, _ = strings.Cut(addInfo, "Lektier:")
			homework = strings.TrimSpace(homework)
			homework = strings.TrimSuffix(homework, "[...]")
		}

		for i, line := range lines {
			if strings.Contains(line, "Hold: ") {
				_, title, _ = strings.Cut(line, ": ")
				title = strings.TrimSpace(title)
				continue
			}
			if strings.Contains(line, "Lærer: ") {
				_, teacher, _ = strings.Cut(line, ": ")
				teacher = strings.TrimSpace(teacher)
				continue
			}
			if strings.Contains(line, "Lokale: ") {
				_, room, _ = strings.Cut(line, ": ")
				room = strings.TrimSpace(room)
				continue
			}

			if i == 0 && (strings.Contains(line, "Ændret!") || strings.Contains(line, "Aflyst!")) {
				status = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(line), "!"))
				fmt.Println(status)
			}

			if strings.Contains(line, "til") {
				parts := strings.Split(line, "til")
				if len(parts) == 2 {
					startDateTime, err := time.ParseInLocation("02/1-2006 15:04", strings.TrimSpace(parts[0]), location)
					// fmt.Println(startDateTime)
					if err == nil {
						startDate = startDateTime
						fmt.Println("START TIME", startDate)

					}
					endDateTime, err := time.ParseInLocation("15:04", strings.TrimSpace(parts[1]), location)
					if err == nil {
						fmt.Println("END TIME", endDateTime)
						endDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), endDateTime.Hour(), endDateTime.Minute(), 0, 0, location)
					}
				}
			}

		}
		module := Module{
			Title:     title,
			StartDate: startDate,
			EndDate:   endDate,
			Room:      room,
			Teacher:   teacher,
			Homework:  homework,
			Status:    status,
		}
		modules = append(modules, module)
	})

	weekString := fmt.Sprintf("%v%v", week, time.Now().Year())
	scheduleUrl := fmt.Sprintf("https://www.lectio.dk/lectio/143/SkemaNy.aspx?week=%v", weekString)
	c.Visit(scheduleUrl)

	wg.Wait()
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].StartDate.Before(modules[j].StartDate)
	})
	return modules
}

func AddToGoogleCalendar(modules []Module, calendarID string) {
	ctx := context.Background()
	bytes, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Could not read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(bytes, calendar.CalendarEventsScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := googlecalendar.GetClient(config)

	service, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Could not get Calendar client: %v", err)
	}

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
		event, err := service.Events.Insert(calendarID, moduleEvent).Do()
		if err != nil {
			log.Fatalf("Unable to create event. %v\n", err)
		}
		fmt.Printf("Created event: %s\n", event.HtmlLink)
	}

	fmt.Println("Added modules to Calendar")
}
