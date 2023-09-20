package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
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
	Id           string    `json:"id"`        // The ID of the module
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
	//LoginInfo *LectioLoginInfo
}

func NewLectio(loginInfo *LectioLoginInfo) *Lectio {
	fmt.Println(loginInfo)
	// lectio.LoginInfo = loginInfo
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
		//LoginInfo: loginInfo,
	}
}

func (*Lectio) GetSchedule(c *colly.Collector, week int) (modules map[string]Module, err error) {
	// var err error
	// var wg sync.WaitGroup
	startTime := time.Now()
	defer fmt.Printf("Took %v\n", time.Since(startTime))
	modules = make(map[string]Module)
	// mu := sync.Mutex{}

	c.OnHTML("a.s2skemabrik.s2brik", func(e *colly.HTMLElement) {
		// wg.Add(1)
		// go func() {

		addInfo := e.Attr("data-additionalinfo")

		lines := strings.Split(addInfo, "\n")

		var id, title, teacher, room, homework string

		// Get ID of the module
		idUrl, _ := url.Parse(e.Attr("href"))
		urlParams, _ := url.ParseQuery(idUrl.RawQuery)
		if strings.Contains(idUrl.RawQuery, "absid") {
			id = urlParams.Get("absid")
		}
		if strings.Contains(idUrl.RawQuery, "aftaleid") {
			id = urlParams.Get("aftaleid")
			e.ForEach("div.s2skemabrikcontent > span", func(i int, e *colly.HTMLElement) {
				title = e.Text
			})

		}

		var status = "uændret"
		var startDate, endDate time.Time

		if strings.Contains(addInfo, "Lektier:") {
			_, homework, _ = strings.Cut(addInfo, "Lektier:")
			homework = strings.TrimSpace(homework)
			homework = strings.TrimSuffix(homework, "[...]")
		}

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

			if strings.Contains(line, "til") && startDate.IsZero() && endDate.IsZero() {

				var convErr error
				startDate, endDate, convErr = ConvertLectioDate(line)

				if convErr != nil {
					log.Printf("Could not convert string to date: %v\n", err)
				}
				continue

			}

		}

		module := Module{
			Id:           id,
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
	// fmt.Println(modules)
	return modules, nil

}

func (l *Lectio) GetScheduleWeeks(weekCount int) (modules map[string]Module, err error) {
	modules = make(map[string]Module)
	_, week := time.Now().ISOWeek()

	for i := 0; i < weekCount; i++ {
		weekModules, err := l.GetSchedule(l.Collector, week+i)
		if err != nil {
			return nil, err
		}
		maps.Copy(modules, weekModules)
	}

	// modules := l.GetSchedule(l.Collector, week)
	// wg.Wait()
	return modules, nil
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

func EventToModule(event *calendar.Event) (module *Module, err error) {
	var teacher, homework, status string
	start, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		log.Fatalf("Could not parse date: %v\n", err)
	}

	end, err := time.Parse(time.RFC3339, event.End.DateTime)
	if err != nil {
		return nil, err
		// log.Fatalf("Could not parse end date: %v\n", err)
	}
	module = &Module{
		Id:           event.Id,
		Title:        event.Summary,
		Room:         event.Location,
		StartDate:    start,
		EndDate:      end,
		Teacher:      teacher,
		Homework:     homework,
		ModuleStatus: status,
	}
	return module, nil
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

func ConvertLectioDate(s string) (startTime time.Time, endTime time.Time, err error) {
	location, err := time.LoadLocation("Europe/Copenhagen")
	if err != nil {
		return startTime, endTime, err
		// log.Fatalf("Could not load location: %v\n", err)
	}
	layout := "2/1-2006 15:04"
	split := strings.Split(s, " til ")
	if len(split) != 2 {
		return startTime, endTime, err
	}

	startTime, err = time.ParseInLocation(layout, split[0], location)
	if err != nil {
		return startTime, endTime, err
	}

	date := startTime.Format("2/1-2006")
	endTime, err = time.ParseInLocation(layout, date+" "+split[1], location)
	if err != nil {
		return startTime, endTime, err
	}

	return startTime, endTime, nil
}
