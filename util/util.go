package util

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
)

// Creates a map consisting of all values from both input maps
func MergeMaps[K comparable, V any](m1 map[K]V, m2 map[K]V) map[K]V {
	merged := make(map[K]V)

	for key, value := range m1 {
		merged[key] = value
	}
	for key, value := range m2 {
		merged[key] = value
	}

	return merged
}

// Compares two maps A and B, and returns two maps consisting of extras and missing from A
func CompareMaps[K comparable, V any](from map[K]V, to map[K]V) (extras map[K]V, missing map[K]V) {
	extras = make(map[K]V)
	missing = make(map[K]V)

	// If key of m1 does not exist in m2, add to the missing map
	for key, value := range from {
		if _, exists := to[key]; !exists {
			missing[key] = value
			// fmt.Printf("ID %v does not exist in to-map\n", key)
		}
	}

	for key, value := range to {
		if _, exists := from[key]; !exists {
			extras[key] = value
			// fmt.Printf("ID %v does not exist in from-map and is extra\n", key)
		}
	}

	return extras, missing
}

// Returns a JSON string representation of a struct
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func ExportSchools(format, outputPath string) error {
	baseURL := "https://lectio.dk/lectio/login_list.aspx"
	c := colly.NewCollector()

	type school struct {
		SchoolID string `json:"schoolID"`
		Name     string `json:"name"`
	}
	var schools []school

	c.OnHTML(".buttonHeader>a[href]", func(h *colly.HTMLElement) {
		link := h.Attr("href")
		if strings.Contains(link, "/default.aspx") {
			schoolName := h.Text
			var schoolID string
			re := regexp.MustCompile(`/lectio/(\d+)/default.aspx`)
			matches := re.FindStringSubmatch(link)

			if len(matches) == 2 {
				schoolID = matches[1]
				schools = append(schools, school{
					SchoolID: schoolID,
					Name:     schoolName,
				})
			}

			return
		}
	})

	err := c.Visit(baseURL)
	if err != nil {
		return err
	}

	for _, school := range schools {
		log.Printf(PrettyPrint(school))
	}
	switch format {
	case "json":
		// Make sure that filename is specified in output path
		if !strings.HasSuffix(outputPath, ".json") {
			outputPath += ".json"
		}
		f, err := os.OpenFile(outputPath, os.O_RDWR | os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		defer f.Close()
		err = json.NewEncoder(f).Encode(schools)
		if err != nil {
			return err
		}
	case "yaml":
		break
	}


	return nil
}
