package util

import (
	"encoding/json"
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
