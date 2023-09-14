package main

import (
	"fmt"
	"lectio-scraper/lectio"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

type Module struct {
	Title     string    `json:"title"`     // Title of the module (eg. 3a Dansk)
	StartDate time.Time `json:"startDate"` // The start date of the module. This includes the date aswell as the time of start (eg. 09:55)
	EndDate   time.Time `json:"endDate"`   // The end date of the module. This includes the date aswell as the time of end (eg. 11:25)
	Room      string    `json:"room"`      // The room of the module (eg. 22)
	Teacher   string    `json:"teacher"`   // The teacher of the class
	Homework  string    `json:"homework"`  // Homework for the module
	Status    string    `json"status"`     // The status of the module (eg. "Ændret" or "Aflyst")
}

var userName string = ""
var password string = ""
var schoolID string = ""

func main() {
	c := colly.NewCollector(colly.AllowedDomains("lectio.dk", "www.lectio.dk"))
	loginUrl := fmt.Sprintf("https://www.lectio.dk/lectio/%s/login.aspx", schoolID)
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	authToken := lectio.GetToken(loginUrl, client)
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
		log.Fatal(err)
	}

	c.OnResponse(func(r *colly.Response) {
		log.Println("response received", r.StatusCode)
	})

	c.OnHTML("div.maintitle", func(e *colly.HTMLElement) {
		fmt.Println(e.Text)
	})

	err = c.Visit("https://www.lectio.dk/lectio/143/forside.aspx")
	if err != nil {
		log.Fatalf("Could not visit %s. %s", "https://www.lectio.dk/lectio/143/forside.aspx", err)
	}

	getSchedule(c)

	// for _, module := range modules {
	// 	fmt.Println(module.Teacher)
	// }
	fmt.Println("Opened login page")
}

func getSchedule(c *colly.Collector) []Module {
	wg := sync.WaitGroup{}
	modules := []Module{}
	c.OnHTML("a.s2skemabrik.s2brik", func(e *colly.HTMLElement) {
		wg.Add(1)
		defer wg.Done()
		addInfo := e.Attr("data-additionalinfo")

		lines := strings.Split(addInfo, "\n")

		var title, teacher, room, status string
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
				status = strings.TrimSpace(line)
			}
		}

		module := Module{
			Title:     title,
			StartDate: time.Now(),
			EndDate:   time.Now(),
			Room:      room,
			Teacher:   teacher,
			Homework:  "",
			Status:    status,
		}
		modules = append(modules, module)
	})
	c.Visit("https://www.lectio.dk/lectio/143/SkemaNy.aspx")
	wg.Wait()
	return modules
}
