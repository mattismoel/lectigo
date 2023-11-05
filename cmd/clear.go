/*
Copyright © 2023 Mattis Møl Kristensen <mattismoel@gmail.com>
*/
package cmd

import (
	"log"
	"os"
	"strings"

	"github.com/mattismoel/lectigo/pkg/lectigo"
	"github.com/mattismoel/lectigo/util"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// clearCmd represents the clear command
var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clears the users Google Calendar",
	Long: `Clears the users Google Calendar from Lectio events. 
	When used, only Lectio events are targeted, therefore leaving any personal events intact.`,
	Run: func(cmd *cobra.Command, args []string) {
		calendarID, err := cmd.Flags().GetString("calendarID")
		if err != nil {
			log.Fatalf("Could not get calendar ID: %v\n", err)
		}
		tokenPath, err := cmd.Flags().GetString("token")
		if err != nil {
			log.Fatalf("Could not get token: %v\n", err)
		}
		// Reads the credentials file and creates a config from it - this is used to create the client
		bytes, err := os.ReadFile("credentials.json")
		if err != nil {
			log.Fatalf("Could not read contents of credentials.json: %v\n", err)
		}

		config, err := google.ConfigFromJSON(bytes, calendar.CalendarEventsScope)
		if err != nil {
			log.Fatalf("Could not create config from credentials.json")
		}

		if !strings.HasSuffix(tokenPath, ".json") {
			tokenPath += ".json"
		}

		client, err := util.GetClient(config, tokenPath)
		if err != nil {
			log.Fatalf("Could not get Google Calendar client: %v\n", err)
		}

		c, err := lectigo.NewGoogleCalendar(client, calendarID)
		if err != nil {
			log.Fatalf("Could not create Google Calendar instance: %v\n", err)
		}

		err = c.Clear()
		if err != nil {
			log.Fatalf("Could not clear Google Calendar: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)

	clearCmd.Flags().StringP("calendarID", "c", "primary", "The Google Calendar ID")
	clearCmd.Flags().StringP("token", "t", "token.json", "The OAuth token file for Google Calendar")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clearCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clearCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
