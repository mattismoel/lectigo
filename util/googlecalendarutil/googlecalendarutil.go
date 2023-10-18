package googlecalendarutil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

// Rerturns a Lectio status based on the color id of a Google Calendar event
func StatusFromColorID(colorId string) string {
	switch colorId {
	case "4":
		return "aflyst"
	case "2":
		return "ændret"
	}
	return "uændret"
}

// Returns a Google Calendar color ID from a Lectio module status
// Aflyst: "4" - red
// Ændret: "2" - green
// Default "" - default calendar color
func ColorIDFromStatus(status string) string {
	switch status {
	case "aflyst":
		return "4"
	case "ændret":
		return "2"
	}
	return ""
}

// Returns the HTTP client from a token.json file, if present
func GetClient(config *oauth2.Config) (*http.Client, error) {
	tokenFile := "token.json"
	token, err := tokenFromFile(tokenFile)
	if err != nil {
		token, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		saveToken(tokenFile, token)
	}
	return config.Client(context.Background(), token), nil
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Printf("Go to the following link in your browser and type the authorization code %q\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, err
	}

	token, err := config.Exchange(context.TODO(), authCode)

	if err != nil {
		return nil, err
	}

	return token, nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return err
	}
	return nil
}
