# Lectio scraper

This project is a web scraper for the Danish gymnasium site Lectio. With this you are able to extract your schedule from the main site, and have it sync with your Google Calendar


# Tools 

The project makes use of the GO programming language


# How to use

Clone the repository to a desired location

```bash
$ git clone "https://github.com/mattismoel/go-lectio-scraper.git
```

Tidy the project

```bash
$ go mod tidy
```


Run the project

```bash
$ go run .
```

In the file `lectio-scraper.go` you will see the different methods for retreiving the modules.