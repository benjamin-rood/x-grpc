package main

import (
	"encoding/json"
	"testing"

	"github.com/go-test/deep"
)

func TestProcessJSON(t *testing.T) {
	// setup: store the test blob in a buffered version of OpenWriteCloserLoader
	buf := NewBufferWriter()
	buf.m["testBlob"] = []byte(jsonBlob)
	cases := []struct {
		testName string
		filename string
		x        OpenWriteCloserLoader
		want     []byte
		err      error
	}{
		{"blob with matching data to modify", "testBlob", buf, []byte(expectedOutput), nil},
	}
	for _, tt := range cases {
		t.Run(tt.testName, func(t *testing.T) {
			err := ProcessJSON(tt.filename, tt.x)
			if err != tt.err {
				t.Error("unexpected error when processing json blob")
			}
			// get resulting modified json
			got, _ := tt.x.Load("modified_" + tt.filename)
			// Compare the modified JSON with the expected output
			// Unmarshal resulting json blob into a map to make comparison easier
			gotDataMap := map[string]any{}
			if err := json.Unmarshal(got, &gotDataMap); err != nil {
				t.Fatal(err)
			}
			// Unmarshal expectedOutput into a map to make comparison easier
			expectedDataMap := map[string]any{}
			if err := json.Unmarshal(tt.want, &expectedDataMap); err != nil {
				t.Fatal(err)
			}
			if diff := deep.Equal(gotDataMap, expectedDataMap); diff != nil {
				t.Errorf("%v", diff)
			}
		})
	}
}
