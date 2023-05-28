package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"
)

func TestProcessJSON(t *testing.T) {

}

// valueEqual checks if two values are equal
func valueEqual(a, b interface{}) bool {
	// log.Printf("a=%v (%s), b=%v (%s)\n", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
	switch a := a.(type) {
	case map[string]interface{}:
		if b, ok := b.(map[string]interface{}); ok {
			return mapEqual(a, b)
		}
	case []interface{}:
		if b, ok := b.([]interface{}); ok {
			return sliceEqual(a, b)
		}
	default:
		return a == b
	}

	return false
}

// mapEqual checks if two maps are equal
func mapEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for k, va := range a {
		vb, ok := b[k]
		if !ok || !valueEqual(va, vb) {
			return false
		}
	}

	return true
}

// sliceEqual checks if two slices are equal
func sliceEqual(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !valueEqual(a[i], b[i]) {
			return false
		}
	}

	return true
}

// marshalWithSortedKeys marshals data to JSON with sorted keys
// used for visual debugging & comparing stringified json
func marshalWithSortedKeys(data any) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	var jsonObj map[string]any
	if err := json.Unmarshal(jsonBytes, &jsonObj); err != nil {
		panic(err)
	}

	sortedKeys := make([]string, 0, len(jsonObj))
	for key := range jsonObj {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	result := "{\n"
	for i, key := range sortedKeys {
		value, _ := json.Marshal(jsonObj[key])
		result += fmt.Sprintf("\t\"%s\": %s", key, value)
		if i < len(sortedKeys)-1 {
			result += ",\n"
		}
	}
	result += "\n}"

	return result
}
