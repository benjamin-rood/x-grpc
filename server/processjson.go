package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

/*
* Bonus Requirements instructions:
* After the content is written to a file on the server the attempt is made to parse it as JSON data
* If the JSON unmarshalling succeeds then the following modifications are done in the JSON data:

  - The properties that start with a vowel should be removed from the JSON data

  - The properties that have even integer number should be increased by *1000*

  - It is expected that the corresponding automated test coverage is included

    Ben's note: since the saved file has no guaranteed file size limit* (hypothetically it could
    be greater than the availability of the available memory), the only way to prevent
    crashing by running out of memory would be to do an on-disk byte traversal of the JSON tree.
    Since this is not part of the assignment, we just error out if the JSON file to be loaded exceeds
    available memory.
*/
func ProcessJSON(filename string, x OpenWriteCloserLoader) error {
	defer x.Close()
	// open file again, and load it all in to memory,
	// (calls Open with the `currentFilename` to open the same file)
	fileContent, err := x.Load(filename)
	if err != nil {
		return err
	}
	// make changes described in bonus requirements
	modifiedData, err := modifyJSON(fileContent)
	if err != nil {
		return err
	}
	// write file contents with modified JSON data to a new file
	if err := x.Open("modified_" + filename); err != nil {
		return err
	}
	if _, err := x.Write(modifiedData); err != nil {
		return fmt.Errorf("failed to write modified JSON data to file: %w", err)
	}
	return nil
}

func modifyJSON(data []byte) ([]byte, error) {
	// unmarshal into a working map where we will do the modifications
	// before marshalling back into a JSON byte slice
	var x map[string]any
	// unmarshal numerical values to `json.Number` instead of float64
	if err := numericalUnmarshalJSON(data, &x); err != nil {
		return nil, fmt.Errorf("not valid json data")
	}
	// Remove keys starting with a vowel
	removeKeysWithVowelPrefix(x)
	// Multiply even integers by 1000
	multiplyEvenIntegers(x)
	// spit back out JSON byte slice
	return json.Marshal(x)
}

// unmarshals numerical values to `json.Number` instead of float64
func numericalUnmarshalJSON(data []byte, v any) error {
	reader := bytes.NewReader(data)
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()

	if err := decoder.Decode(v); err != nil {
		return err
	}

	return nil
}

func removeKeysWithVowelPrefix(data map[string]any) {
	vowels := []string{"a", "e", "i", "o", "u"}
	keysToRemove := []string{}

	for key, value := range data {
		// Check if the key starts with a vowel
		firstChar := strings.ToLower(key[0:1])
		for _, vowel := range vowels {
			if firstChar == vowel {
				keysToRemove = append(keysToRemove, key)
				break
			}
		}

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
		switch val := value.(type) {
		case json.Number:
			if intVal, err := strconv.ParseInt(val.String(), 10, 64); err == nil {
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
		switch val := item.(type) {
		case json.Number:
			if intVal, err := strconv.ParseInt(val.String(), 10, 64); err == nil {
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
