package main

import (
	"encoding/json"
	"fmt"
	"time"
)

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

func CompareMaps[K comparable, V any](from map[K]V, to map[K]V) (extras map[K]V, missing map[K]V) {
	extras = make(map[K]V)
	missing = make(map[K]V)

	// If key of m1 does not exist in m2, add to the missing map
	for key, value := range from {
		if _, exists := to[key]; !exists {
			missing[key] = value
			fmt.Printf("ID %v does not exist in to-map\n", key)
		}
	}

	for key, value := range to {
		if _, exists := from[key]; !exists {
			extras[key] = value
			fmt.Printf("ID %v does not exist in from-map and is extra\n", key)
		}
	}

	return extras, missing
}

func RoundDateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func GetMonday() time.Time {
	t := time.Now()
	offset := int(time.Monday - t.Weekday())
	if offset > 0 {
		offset = -6
	}
	return t.AddDate(0, 0, offset)
}
