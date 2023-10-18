package googlecalendar

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mattismoel/lectigo/types"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Creates a new Google Calendar struct instance
func New(client *http.Client, calendarID string) (*types.GoogleCalendar, error) {
	ctx := context.Background()

	service, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	
	calendar := &types.GoogleCalendar{
		Service: service,
		ID:      calendarID,
		Logger:  log.New(os.Stdout, "google-calendar ", log.LstdFlags),
	}
	return calendar, nil
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
		log.Fatalf("Could not read authorization code: %v\n", err)
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
