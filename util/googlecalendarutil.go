package util

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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
func GetClient(config *oauth2.Config, tokenPath string) (*http.Client, error) {
	token, err := tokenFromFile(tokenPath)
	if err != nil {
		token, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		saveToken(tokenPath, token)
	}
	return config.Client(context.Background(), token), nil
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	var authCode string

	fmt.Printf("Visit here: %q\n", authURL)

	server := &http.Server{Addr: ":8080"}
	http.HandleFunc("/oauth", func(w http.ResponseWriter, r *http.Request) {
		authCode = r.URL.Query().Get("code")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("could not shut down server: %v\n", err)
		}

	})

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
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
