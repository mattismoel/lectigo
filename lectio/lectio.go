package lectio

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/mattismoel/lectigo/types"
)

func New(loginInfo *types.LectioLoginInfo) *types.Lectio {
	loginUrl := fmt.Sprintf("https://www.lectio.dk/lectio/%s/login.aspx", loginInfo.SchoolID)
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	collector := colly.NewCollector(colly.AllowedDomains("lectio.dk", "www.lectio.dk"))
	authToken := types.GetToken(loginUrl, client)

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

	return &types.Lectio{
		Client:    client,
		Collector: collector,
		//LoginInfo: loginInfo,
	}
}

func ModulesToJSON(modules map[string]types.Module, filename string) {
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
