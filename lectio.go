package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"golang.org/x/exp/maps"
	"google.golang.org/api/calendar/v3"
)

type App struct {
	Client *http.Client
}

type AuthenticityToken struct {
	Token string
}

type LectioLoginInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
	SchoolID string `json:"schoolID"`
}

type Module struct {
	Title        string    `json:"title"`     // Title of the module (eg. 3a Dansk)
	StartDate    time.Time `json:"startDate"` // The start date of the module. This includes the date as well as the time of start (eg. 09:55)
	EndDate      time.Time `json:"endDate"`   // The end date of the module. This includes the date as well as the time of end (eg. 11:25)
	Room         string    `json:"room"`      // The room of the module (eg. 22)
	Teacher      string    `json:"teacher"`   // The teacher of the class
	Homework     string    `json:"homework"`  // Homework for the module
	ModuleStatus string    `json:"status"`    // The status of the module (eg. "Ændret" or "Aflyst")
}

type Lectio struct {
	Client    *http.Client
	Collector *colly.Collector
}

func NewLectio(loginInfo *LectioLoginInfo) *Lectio {
	fmt.Println(loginInfo)
	loginUrl := fmt.Sprintf("https://www.lectio.dk/lectio/%s/login.aspx", loginInfo.SchoolID)
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	collector := colly.NewCollector(colly.AllowedDomains("lectio.dk", "www.lectio.dk"))

	authToken := GetToken(loginUrl, client)

	// Attempts to log the user in with the given login information
	err := collector.Post(loginUrl, map[string]string{
		"m$Content$username": loginInfo.Username,
		"m$Content$password": loginInfo.Password,
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
	collector.OnResponse(func(r *colly.Response) {
		log.Println("response received", r.StatusCode, r.Request.URL)
	})

	return &Lectio{
		Client:    client,
		Collector: collector,
	}
}

func (*Lectio) GetSchedule(c *colly.Collector, week int) map[string]Module {
	startTime := time.Now()
	defer fmt.Printf("GetSchedule took %v\n", time.Since(startTime))
	modules := make(map[string]Module)

	c.OnHTML("a.s2skemabrik.s2brik", func(e *colly.HTMLElement) {
		// Gets the string from the attribute "data-additionalinfo" which contains all nescessary information about the module
		//
		// Example string:
		// ============================
		// Ændret!
		// 18/9-2023 12:00 til 13:30
		// Hold: 3a HI
		// Lærer: <Teacher Name> (<Teacher Acronym>)
		// Lokale: <Room Number/Room Name>
		// Lektier:
		// <Homework>
		// ============================
		addInfo := e.Attr("data-additionalinfo")

		// Creates a slice where each entry is a line of the addInfo
		lines := strings.Split(addInfo, "\n")

		var id, title, teacher, room, homework string

		// Get ID of the module - this is used to sync Lectios schedule with the users Google Calendar events
		idUrl, _ := url.Parse(e.Attr("href"))
		urlParams, _ := url.ParseQuery(idUrl.RawQuery)

		// Checks which type of event the current entry is - this can be a private appointment or a class
		if strings.Contains(idUrl.RawQuery, "absid") {
			id = urlParams.Get("absid")
		}
		if strings.Contains(idUrl.RawQuery, "aftaleid") {
			id = urlParams.Get("aftaleid")
			e.ForEach("div.s2skemabrikcontent > span", func(i int, e *colly.HTMLElement) {
				title = e.Text
			})

		}

		var status = "uændret"           // Sets the default status of the module
		var startDate, endDate time.Time // Initialises start and end dates - these have zero values upon initialisation, which will be important when parsing the lines later

		// Checks for homework string and assigns the variable to a formatted string
		if strings.Contains(addInfo, "Lektier:") {
			_, homework, _ = strings.Cut(addInfo, "Lektier:")
			homework = strings.TrimSpace(homework)
			homework = strings.TrimSuffix(homework, "[...]")
		}

		// Cycles through each line of the attribute string and checks assigns the variables their values
		for i, line := range lines {
			if strings.Contains(line, "Hold: ") && title == "" {
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
				continue
			}

			// Checks if the line is the one that contains the start and end date of the module.
			// The variables will only be assigned as long as the start and end date have not been initialised yet
			if strings.Contains(line, "til") && startDate.IsZero() && endDate.IsZero() {
				startDate, endDate, _ = ConvertLectioDate(line)
				continue
			}

		}

		// Initialises a variable that contains the fitting data and adds it to the return map
		module := Module{
			Title:        title,
			StartDate:    startDate,
			EndDate:      endDate,
			Room:         room,
			Teacher:      teacher,
			Homework:     homework,
			ModuleStatus: status,
		}
		modules[id] = module
	})

	weekString := fmt.Sprintf("%v%v", week, time.Now().Year())
	scheduleUrl := fmt.Sprintf("https://www.lectio.dk/lectio/143/SkemaNy.aspx?week=%v", weekString)
	c.Visit(scheduleUrl)
	return modules
}

func (l *Lectio) GetScheduleWeeks(weekCount int) map[string]Module {
	modules := make(map[string]Module)
	_, week := time.Now().ISOWeek()

	for i := 0; i < weekCount; i++ {
		weekModules := l.GetSchedule(l.Collector, week+i)
		maps.Copy(modules, weekModules)
	}

	return modules
}

func GetToken(loginUrl string, client *http.Client) AuthenticityToken {
	response, err := client.Get(loginUrl)

	if err != nil {
		log.Fatal("Error fetching response: ", err)
	}

	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body.", err)
	}

	token, _ := document.Find("input[name=__EVENTVALIDATION]").Attr("value")

	authenticityToken := AuthenticityToken{Token: token}
	return authenticityToken
}

func EventToModule(event *calendar.Event) Module {
	var teacher, homework, status string
	start, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		log.Fatalf("Could not parse date: %v\n", err)
	}

	end, err := time.Parse(time.RFC3339, event.End.DateTime)
	if err != nil {
		log.Fatalf("Could not parse end date: %v\n", err)
	}
	return Module{
		Title:        event.Summary,
		Room:         event.Location,
		StartDate:    start,
		EndDate:      end,
		Teacher:      teacher,
		Homework:     homework,
		ModuleStatus: status,
	}
}

func ModulesToJSON(modules map[string]Module, filename string) {
	filename, _ = strings.CutSuffix(filename, ".json")
	b, err := json.Marshal(modules)
	if err != nil {
		log.Fatal("Could not marshal JSON.", err)
	}

	err = os.WriteFile(fmt.Sprintf("%s.json", filename), b, 0644)
	if err != nil {
		log.Fatal("Could not write to file", err)
	}
}

func ConvertLectioDate(input string) (time.Time, time.Time, error) {
	re := regexp.MustCompile(`(\d{1,2}/\d{1,2}-\d{4} \d{2}:\d{2}) til (\d{2}:\d{2})`)
	match := re.FindStringSubmatch(input)
	if len(match) > 0 {
		location, err := time.LoadLocation("Europe/Copenhagen")
		if err != nil {
			log.Fatalf("Could not load location %v\n", "Europe/Copenhagen")
		}
		dateParts := strings.Split(match[1], " ")
		startDateStr := match[1]
		endDateStr := dateParts[0] + " " + match[2]

		layout := "2/1-2006 15:04"
		startDate, err1 := time.ParseInLocation(layout, startDateStr, location)
		endDate, err2 := time.ParseInLocation(layout, endDateStr, location)

		if err1 != nil {
			return time.Time{}, time.Time{}, err1
		}
		if err2 != nil {
			return time.Time{}, time.Time{}, err2
		}

		return startDate, endDate, nil
	}

	return time.Time{}, time.Time{}, fmt.Errorf("no date found")
}
