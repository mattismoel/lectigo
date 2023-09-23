package lectigo

// The login information of the user. This should be stored in a "secrets.json" file, and should have the following variables: username, password, schoolID, calendarID
// var lectioLoginInfo *LectioLoginInfo

// func main() {
// 	var username, password, schoolID string
// 	var envVarExists bool
// 	var googleCalendarID string

// 	bytes, err := os.ReadFile("credentials.json")
// 	if err != nil {
// 		log.Fatalf("Could not read client secret file: %v", err)
// 	}

// 	config, err := google.ConfigFromJSON(bytes, calendar.CalendarScope)

// 	config.RedirectURL = "http://localhost:3000/oauth/token"

// 	if err != nil {
// 		log.Fatalf("Unable to parse client secret file to config: %v", err)
// 	}

// 	client := *GetClient(config)

// 	err = godotenv.Load(".env")
// 	if err != nil {
// 		log.Fatalf("Could not load .env file: %v\n", err)
// 	}
// 	// Checks if Lectio environment variables exist, and assigns their respecitve values to them
// 	if username, envVarExists = os.LookupEnv("LECTIO_USERNAME"); !envVarExists {
// 		log.Fatalf("Could not get the Lectio username from .env file. Please make sure that it is present\n")
// 	}
// 	if password, envVarExists = os.LookupEnv("LECTIO_PASSWORD"); !envVarExists {
// 		log.Fatalf("Could not get the lectio password from .env file. Please make sure that it is present\n")
// 	}
// 	if schoolID, envVarExists = os.LookupEnv("LECTIO_SCHOOL_ID"); !envVarExists {
// 		log.Fatalf("Could not get the lectio password from .env file. Please make sure that it is present\n")
// 	}

// 	lectioLoginInfo = &LectioLoginInfo{
// 		Username: username,
// 		Password: password,
// 		SchoolID: schoolID,
// 	}

// 	if googleCalendarID, envVarExists = os.LookupEnv("GOOGLE_CALENDAR_ID"); !envVarExists {
// 		log.Fatalf("Could not get the Google Calendar ID from .env file. Please make sure that it is present")
// 	}

// 	var l *Lectio
// 	var c *GoogleCalendar

// 	// Creates command line flags and parses them upon execution of application
// 	command := flag.String("command", "sync", "what command should be executed")
// 	weekCount := flag.Int("weeks", 2, "define amount of weeks to sync Google Calendar")
// 	flag.Parse()

// 	// Checks for the provided commands and creates instances of Lectio and GoogleCalendar only if nescessary
// 	switch *command {
// 	// If user wants to sync Lectio schedule with Google Calendar
// 	// Creates appropriate clients, fetches modules and updates the Google Calendar
// 	case "sync":
// 		log.Printf("Syncing Google Calendar for the next %v weeks...\n", *weekCount)
// 		l = NewLectio(lectioLoginInfo)                       // Creates a new Lectio client
// 		c = NewGoogleCalendar(&client, googleCalendarID)     // Creates a new Google Calendar client
// 		lectioModules, err := l.GetScheduleWeeks(*weekCount) // Gets the modules from the Lectio schedule
// 		if err != nil {
// 			log.Fatalf("Could not get the weekly schedule: %v\n", err)
// 		}
// 		googleModules, err := c.GetModules(*weekCount) // Gets the modules present in Google Calendar
// 		if err != nil {
// 			log.Fatalf("Could not get modules from Google Calendar: %v\n", err)
// 		}
// 		err = c.UpdateCalendar(lectioModules, googleModules) // Updates and deletes events that are missing in Google Calendar
// 		if err != nil {
// 			log.Fatalf("Could not update Google Calendar: %v\n", err)
// 		}
// 	// If user wants to clear the Google Calendar
// 	// Create Google Calendar client and clear the calendar
// 	case "clear":
// 		log.Printf("Clearing Google Calendar...\n")
// 		c = NewGoogleCalendar(&client, googleCalendarID) // Creates a new Google Calendar client
// 		err := c.Clear()
// 		if err != nil {
// 			log.Fatalf("Could not clear the Google Calendar: %v\n", err)
// 		}
// 	}
// }
