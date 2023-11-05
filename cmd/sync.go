/*
Copyright © 2023 Mattis Kristensen <mattismoel@gmail.com>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mattismoel/lectigo/pkg/lectigo"
	"github.com/mattismoel/lectigo/util"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// var (
// 	username   string
// 	password   string
// 	schoolID   string
// 	calendarID string
// 	tokenPath  string
// 	weeks      int
// )

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs a Lectio schedule with a Google Calendar",
	Long:  `Synchronises a users Lectio scedule with Google Calendar. The users Lectio login info as well as Google Calendar info is provided.`,
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		schoolID, _ := cmd.Flags().GetString("schoolID")
		calendarID, _ := cmd.Flags().GetString("calendarID")
		tokenPath, _ := cmd.Flags().GetString("tokenPath")
		weeks, _ := cmd.Flags().GetInt("weeks")

		fmt.Println("Attempting to sync Lectio and Google Calendar...")

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
		l, err := lectigo.NewLectio(&lectigo.LectioLoginInfo{
			Username: username,
			Password: password,
			SchoolID: schoolID,
		})
		if err != nil {
			log.Fatalf("Could not create Lectio instance: %v\n", err)
		}

		lModules, err := l.GetScheduleWeeks(weeks)
		if err != nil {
			log.Fatalf("Could not get Lectio schedule: %v\n", err)
		}

		gEvents, err := c.GetEvents(weeks)
		if err != nil {
			log.Fatalf("Could not get events from Google Calendar: %v\n", err)
		}
		err = c.UpdateCalendar(lModules, gEvents)
		if err != nil {
			log.Fatalf("Could not update Google Calendar: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringP("username", "u", "", "Lectio username (required)")
	syncCmd.Flags().StringP("password", "p", "", "Lectio password (required)")
	syncCmd.Flags().StringP("schoolID", "s", "", "Lectio school ID (required)")
	syncCmd.Flags().IntP("weeks", "w", 2, "Amount of weeks to sync")
	syncCmd.Flags().StringP("calendarID", "c", "primary", "Google Calendar calendar ID")
	syncCmd.Flags().StringP("tokenPath", "t", "token.json", "The path to a Google OAuth token file")

	syncCmd.MarkFlagRequired("username")
	syncCmd.MarkFlagRequired("password")
	syncCmd.MarkFlagRequired("schoolID")
}
