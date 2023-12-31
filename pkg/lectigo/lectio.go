package lectigo

import (
	"encoding/json"
	"errors"
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
	LoginInfo *LectioLoginInfo
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

type AuthenticityToken string

// Creates a new instance of a Lectio struct. Generates a token, if not present in root directory.
func NewLectio(loginInfo *LectioLoginInfo) (*Lectio, error) {
	loginUrl := fmt.Sprintf("https://www.lectio.dk/lectio/%s/login.aspx", loginInfo.SchoolID)
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Jar: jar}
	collector := colly.NewCollector(colly.AllowedDomains("lectio.dk", "www.lectio.dk"))

	authToken, err := GetToken(loginUrl, client)
	if err != nil {
		return nil, err
	}

	// Attempts to log the user in with the given login information
	err = collector.Post(loginUrl, map[string]string{
		"m$Content$username": loginInfo.Username,
		"m$Content$password": loginInfo.Password,
		"__EVENTVALIDATION":  string(*authToken),
		"__EVENTTARGET":      "m$Content$submitbtn2",
		"__EVENTARGUMENT":    "",
		"masterfootervalue":  "X1!ÆØÅ",
		"LectioPostbackId":   "",
	})

	if err != nil {
		return nil, err
	}

	lectio := &Lectio{
		Client:    client,
		Collector: collector,
		LoginInfo: loginInfo,
	}
	return lectio, nil
}

// Converts a Lectio module to a Google Calendar event
func (m *Module) ToGoogleEvent() *GoogleEvent {
	calendarColorID := ""
	switch m.ModuleStatus {
	case "aflyst":
		calendarColorID = "4"
	case "ændret":
		calendarColorID = "2"
	}

	descLayout := `
Lærer: %s
Lektier:
%s
	`
	description := fmt.Sprintf(strings.TrimSpace(descLayout), m.Teacher, m.Homework)
	return &GoogleEvent{
		Id:          "lec" + m.Id,
		Description: description,
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
	}
}

// Gets the Lectio schedule of a specified week number.
func (l *Lectio) GetSchedule(week int) (map[string]Module, error) {
	startTime := time.Now()
	modules := make(map[string]Module)

	// Handle redirects. The collector is not expected to get redirected. If it does, it checks for errors - for example the school id does not exist.
	l.Collector.RedirectHandler = func(req *http.Request, via []*http.Request) error {
		if strings.Contains(req.URL.String(), "fejlhandled") {
			return errors.New("Could not get Lectio schedule. The school ID provided does not exist")
		}
		return nil
	}

	l.Collector.OnHTML("table.s2skema>tbody", func(h *colly.HTMLElement) {
		var weekStart time.Time

		h.ForEach("tr", func(row int, h *colly.HTMLElement) {
			// If on day row
			if row == 1 {
				var err error
				weekStartString := h.ChildText("td:nth-child(2)")
				weekStart, err = parseDate(weekStartString)
				if err != nil {
					log.Fatalf("Could not parse date: %v\n", err)
				}

				h.ForEach("td", func(col int, h *colly.HTMLElement) {
					// If sidebar column return prematurely
					if col == 0 {
						return
					}
				})
			}

			// If on module row
			if row == 3 {
				// Foreach column in schedule
				h.ForEach("td", func(col int, h *colly.HTMLElement) {
					// If sidebar column return prematurely
					if col == 0 {
						return
					}

					// Gets the current date of the module
					date := weekStart.AddDate(0, 0, col-1)

					// Foreach module
					h.ForEach("a.s2skemabrik", func(i int, e *colly.HTMLElement) {
						addInfo := e.Attr("data-additionalinfo")
						if addInfo == "" {
							log.Printf("could not get Lectio schedule for the week %d", week)
						}

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
						var startDate time.Time
						var duration time.Duration

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

							if strings.Contains(line, "-") && startDate.IsZero() && duration == 0 {
								startDate, duration, _ = util.ConvertTimestamp(&date, line)
								continue
							}
						}

						module := Module{
							Id:           id,
							Title:        title,
							StartDate:    startDate,
							EndDate:      startDate.Add(duration),
							Room:         room,
							Teacher:      teacher,
							Homework:     homework,
							ModuleStatus: status,
						}
						modules[id] = module
					})
				})
			}
		})
	})

	weekString := fmt.Sprintf("%v%v", week, time.Now().Year())
	scheduleUrl := fmt.Sprintf("https://www.lectio.dk/lectio/%s/SkemaNy.aspx?week=%v", l.LoginInfo.SchoolID, weekString)
	err := l.Collector.Visit(scheduleUrl)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Got Lectio schedule for week %v in %v\n", week, time.Since(startTime))
	return modules, nil
}

// Parses a Lectio date of format "Onsdag (9/12)" to a time.Time struct
func parseDate(input string) (time.Time, error) {
	datePattern := `\((\d+)/(\d+)\)`
	re := regexp.MustCompile(datePattern)

	match := re.FindStringSubmatch(input)
	if len(match) != 3 {
		return time.Time{}, errors.New("Date not found in the input string")
	}

	currentYear := time.Now().Year()

	day, _ := time.Parse("2/1", match[1]+"/"+match[2])
	date := time.Date(currentYear, day.Month(), day.Day(), 0, 0, 0, 0, time.Local)

	return date, nil
}

// Gets the Lectio schedule from the current weeks and weekCount weeks ahead.
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

func GetToken(loginUrl string, client *http.Client) (*AuthenticityToken, error) {
	response, err := client.Get(loginUrl)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	token, _ := document.Find("input[name=__EVENTVALIDATION]").Attr("value")

	authenticityToken := AuthenticityToken(token)
	return &authenticityToken, nil
}

// Checks if two Lectio modules are equal
func (m1 *Module) Equals(m2 *Module) bool {
	b := m1.Id == m2.Id &&
		m1.StartDate.Equal(m2.StartDate) &&
		m1.EndDate.Equal(m2.EndDate) &&
		m1.ModuleStatus == m2.ModuleStatus &&
		m1.Room == m2.Room
		// m1.Homework != m2.Homework

	return b
}

// Converts input Lectio modules to a JSON object at the specified path
func ModulesToJSON(modules map[string]Module, filename string) error {
	filename, _ = strings.CutSuffix(filename, ".json")
	b, err := json.Marshal(modules)
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s.json", filename), b, 0644)
	if err != nil {
		return err
	}

	return nil
}
