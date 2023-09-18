# Lectio scraper

This project is a web scraper for the Danish gymnasium site Lectio. With this you are able to extract the schedule from the main site, and have it sync with your Google Calendar.


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

You can build the project and run it with the following commands:
```bash
$ go build .
```

## Usage

The project can be run either by the executable or the `go run` command. Generally the executable will be faster.

### Examples 

Syncing Lectio schedule with Google Calendar with modules for the next three weeks:

```bash
$ ./go-lectio-scraper --command=sync --weeks=3
```


Clearing all Lectio modules from Google Calendar
> Note: This DOES NOT delete normal events from your calendar. Only Lectio modules are targeted.

```bash
$ ./go-lectio-scraper --command=clear
```


## Setting up login details

### Google OAuth authentication

This project makes use of the [Google Calendar API](google.golang.org/api/calendar/v3), and therefore needs you to log in with your Google Account. When the application is run for the first time, a link will appear for you to log in. This will return a link in the address bar. Copy the value of `code=` (until `&scope`) and paste it in the terminal and press `Enter`. A token should be generated as `token.json`.


### Lectio login details

The login details of Lectio should be provided in a `lectioSecrets.json` file with the following format:

```json
{
    "username": "<username>",
    "password": "<password>",
    "schoolID": "<school_id>"
}
```

The `schoolID` can be found when visiting the Lectio frontpage - either before or after login at `https://lectio.dk/lectio/<schoolID>/...` 


### Google Calendar ID

If you do not wish to use the default calendar, a `googleSecrets.json` file should be provided. This should have the following formatting:

```json
{
    "calendarID": "<calendar_id>"
}
```

The list of calendars that you own can be found at [Google Calendar API's documentation CalendarList:list()](https://developers.google.com/calendar/api/v3/reference/calendarList/list). In the sidebar you should see an `Execute` button. The resulting JSON is a list of your Google calendars. Find the desired calendar and copy the value of `"id"` and paste it in the `googleSecrets.json` file. Everything should be ready to go.