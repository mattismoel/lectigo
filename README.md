# What is this?

This project is a web scraper for the Danish gymnasium site [Lectio](https://lectio.dk) by MaCom A/S. With this, you are able to extract the schedule from the main site and have it sync with your Google Calendar.


# How to use

Clone the repository to a desired location

```bash
$ git clone "https://github.com/mattismoel/go-lectio-scraper.git
```

Tidy the project

```bash
$ go mod tidy
```

## Building the project

The repository does not contain the executables for the application. These can be built using the following Makefile command:

```bash
$ make build
```

This creates a directory called `bin` wherein all the binaries can be found for MacOS, Linux, and Windows.

## Usage

The project can be run either by the executable (with `make run` or direct call on the executable in the `bin` directory) or the `go run` command. Generally, the executable will be faster.

### Examples 

Syncing Lectio schedule with Google Calendar with modules for the next three weeks:

```bash
$ make run COMMAND=sync WEEKS=3
```
OR
```
./bin/lectigo-<target> --command=sync --weeks=3
```

Clearing all Lectio modules from Google Calendar
> Note: This DOES NOT delete normal events from your calendar. Only Lectio modules are targeted.

```bash
$ make run COMMAND=clear
```
OR
```bash
$ ./bin/lectigo-<target> --command=clear
```
## Command line flags

| Flag      | Purpose                                                                                                                         |
|-----------|---------------------------------------------------------------------------------------------------------------------------------|
| `command` | Chooses the command to execute. Can be set to `sync` or `clear`. Defaults to `command=sync`                                     | 
| `weeks`   | Sets the number of weeks to sync the Calendar. Only affects `command=sync`. Defaults to `1` (getting the schedule of this week) |

> For examples see [Examples](https://github.com/mattismoel/go-lectio-scraper#examples)

# Setting up login details

### Google OAuth authentication

This project makes use of the [Google Calendar API](google.golang.org/api/calendar/v3), and therefore needs you to log in with your Google Account. When the application is run for the first time, a link will appear for you to log in. This will return a link in the address bar. Copy the value of `code=` (until `&scope`) and paste it into the terminal and press `Enter`. A token should be generated as `token.json`.


### Environment variables

The login details of Lectio and the Google Calendar ID of your choice should be provided in a `.env` file with the following variables:

```
LECTIO_USERNAME=<username>
LECTIO_PASSWORD=<password>
LECTIO_SCHOOL_ID=<school_id>
GOOGLE_CALENDAR_ID=<calendar_id>
```

The `schoolID` can be found when visiting the Lectio frontpage - either before or after login at `https://lectio.dk/lectio/<schoolID>/...` 

The list of calendars that you own can be found at [Google Calendar API's documentation CalendarList:list()](https://developers.google.com/calendar/api/v3/reference/calendarList/list). In the sidebar, you should see an `Execute` button. The resulting JSON is a list of your Google calendars. Find the desired calendar, copy the value of `"id"` and paste it into the `googleSecrets.json` file. Everything should be ready to go.

