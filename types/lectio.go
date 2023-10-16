package types

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/mattismoel/lectigo/util"
	"golang.org/x/exp/maps"
	"google.golang.org/api/calendar/v3"
)

type LectioLoginInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
	SchoolID string `json:"schoolID"`
}

type Lectio struct {
	Client    *http.Client
	Collector *colly.Collector
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

type AuthenticityToken struct {
	Token string
}

func (m *Module) ToGoogleEvent() *GoogleEvent {
	calendarColorID := ""
	switch m.ModuleStatus {
	case "aflyst":
		calendarColorID = "4"
	case "ændret":
		calendarColorID = "2"
	}

	// description := fmt.Sprintf("%s\n%s", m.Teacher, m.Homework)
	return &GoogleEvent{
		event: &calendar.Event{
			Id:          "lec" + m.Id,
			Description: "",
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
			Status:   "confirmed",
		},
	}
}

func (l *Lectio) GetSchedule(week int) (modules map[string]Module, err error) {
	startTime := time.Now()
	defer fmt.Printf("Got Lectio schedule for week %v in %v\n", week, time.Since(startTime))
	modules = make(map[string]Module)

	l.Collector.OnHTML("a.s2skemabrik.s2brik", func(e *colly.HTMLElement) {
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
			if strings.Contains(line, "Lokale: ") || strings.Contains(line, "Lokaler: ") {
				_, room, _ = strings.Cut(line, ": ")
				room = strings.TrimSpace(room)
				continue
			}

			if i == 0 && (strings.Contains(line, "Ændret!") || strings.Contains(line, "Aflyst!")) {
				status = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(line), "!"))
				continue
			}

			if strings.Contains(line, "til") && startDate.IsZero() && endDate.IsZero() {
				startDate, endDate, _ = util.ConvertLectioDate(line)
				continue
			}
		}

		module := Module{
			Id:           id,
			Title:        title,
			StartDate:    startDate,
			EndDate:      endDate,
			Room:         room,
			Teacher:      "",
			Homework:     homework,
			ModuleStatus: status,
		}

		modules[id] = module
	})

	weekString := fmt.Sprintf("%v%v", week, time.Now().Year())
	scheduleUrl := fmt.Sprintf("https://www.lectio.dk/lectio/143/SkemaNy.aspx?week=%v", weekString)
	l.Collector.Visit(scheduleUrl)
	return modules, nil
}

func (l *Lectio) GetScheduleWeeks(weekCount int) (modules map[string]Module, err error) {
	modules = make(map[string]Module)
	_, week := time.Now().ISOWeek()

	for i := 0; i < weekCount; i++ {
		weekModules, err := l.GetSchedule(week + i)
		if err != nil {
			return nil, err
		}
		maps.Copy(modules, weekModules)
	}

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

func (m1 *Module) Equals(m2 *Module) bool {
	b := m1.Id == m2.Id && 
	m1.StartDate.Equal(m2.StartDate) && 
	m1.EndDate.Equal(m2.EndDate) && 
	m1.ModuleStatus == m2.ModuleStatus && 
	m1.Room == m2.Room

	return b
}