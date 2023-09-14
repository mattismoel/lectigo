package utils

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
