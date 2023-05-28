package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"testing"
)

func TestUploaderService(t *testing.T) {
	// define listener
	// initialise grpc server
	// register server with uploader service
	// initialise grpc service client
	// use generated client for testing

}

// quick bootstrap of bytes.Buffer to use for testing
type bufwc struct {
	buffer *bytes.Buffer
}

func (b *bufwc) Open(string) error {
	// ignore filename
	b.buffer = bytes.NewBuffer([]byte{})
	return nil
}

func (b *bufwc) Write(p []byte) (n int, err error) {
	return b.buffer.Write(p)
}

func (b *bufwc) Close() error {
	return nil // Since it's just a buffer, closing has no effect
}

func (b *bufwc) String() string {
	return b.buffer.String()
}

// Check interface conformity
var _ OpenWriteCloser = &bufwc{}

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
// no functional use in the code, just used for visually debugging
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
