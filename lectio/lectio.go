package lectio

import (
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type App struct {
	Client *http.Client
}

type AuthenticityToken struct {
	Token string
}

// func (app *App) Login(loginUrl string, username string, password string) {
// 	client := app.Client

// 	authenticityToken := app.getToken(loginUrl)

// 	data := url.Values{
// 		"m$Content$username": {username},
// 		"m$Content$password": {password},
// 		"__EVENTVALIDATION":  {authenticityToken.Token},
// 		"__EVENTTARGET":      {"m$Content$submitbtn2"},
// 		"__EVENTARGUMENT":    {""},
// 		"masterfootervalue":  {"X1!ÆØÅ"},
// 		"LectioPostbackId":   {""},
// 	}

// 	res, err := client.PostForm(loginUrl, data)

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	defer res.Body.Close()

// 	_, err = ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// }

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
