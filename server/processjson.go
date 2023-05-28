package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

func processJSON(data []byte) ([]byte, error) {
	if !gjson.Valid(string(data)) {
		return nil, fmt.Errorf("not valid json data")
	}
	return nil, fmt.Errorf("not implemented")
}

// unmarshals numerical values to `json.Number` instead of float64
func unmarshalJSON(data []byte, v any) error {
	reader := bytes.NewReader(data)
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()

	if err := decoder.Decode(v); err != nil {
		return err
	}

	return nil
}

func in[T comparable](val T, slice []T) bool {
	for _, s := range slice {
		if val == s {
			return true
		}
	}
	return false
}

func removeKeysWithVowelPrefix(data map[string]any) {
	vowelKeys := []string{"a", "e", "i", "o", "u"}
	keysToRemove := []string{}

	for key, value := range data {
		// Check if the key starts with a vowel
		firstChar := strings.ToLower(key[0:1])
		// for _, vowel := range vowelKeys {
		if in(firstChar, vowelKeys) {
			keysToRemove = append(keysToRemove, key)
			break
		}
		// }

		// Recursive call for nested objects
		if nested, ok := value.(map[string]any); ok {
			removeKeysWithVowelPrefix(nested)
		}

		// Recursive call for arrays
		if arr, ok := value.([]any); ok {
			removeKeysWithVowelPrefixArray(arr)
		}
	}

	// Remove the identified keys
	for _, key := range keysToRemove {
		delete(data, key)
	}
}

func removeKeysWithVowelPrefixArray(arr []any) {
	for _, item := range arr {
		if nested, ok := item.(map[string]any); ok {
			removeKeysWithVowelPrefix(nested)
		}

		if nestedArr, ok := item.([]any); ok {
			removeKeysWithVowelPrefixArray(nestedArr)
		}
	}
}

func multiplyEvenIntegers(data map[string]any) {
	for key, value := range data {
		log.Printf("type: %s\n", reflect.TypeOf(value))
		switch val := value.(type) {
		case json.Number:
			log.Println("Found a Number:", val)
			if intVal, err := strconv.ParseInt(val.String(), 10, 64); err == nil {
				log.Println("Found an integer:", intVal)
				if intVal%2 == 0 {
					// Multiply even integer values by 1000
					data[key] = json.Number(strconv.FormatInt(intVal*1000, 10))
				}
			}

		case []any:
			multiplyEvenIntegersArray(val)
		case map[string]any:
			multiplyEvenIntegers(val)
		}
	}
}

func multiplyEvenIntegersArray(arr []any) {
	for i, item := range arr {
		log.Printf("type: %s\n", reflect.TypeOf(item))
		switch val := item.(type) {
		case json.Number:
			log.Println("Found a Number:", val)
			if intVal, err := strconv.ParseInt(val.String(), 10, 64); err == nil {
				log.Println("Found an integer:", intVal)
				if intVal%2 == 0 {
					// Multiply even integer values by 1000
					arr[i] = json.Number(strconv.FormatInt(intVal*1000, 10))
				}
			}
		case []any:
			multiplyEvenIntegersArray(val)
		case map[string]any:
			multiplyEvenIntegers(val)
		}
	}
}
