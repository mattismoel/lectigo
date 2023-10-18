package googlecalendarutil

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

