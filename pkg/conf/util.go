package conf

import (
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/jsonx"
)

// UnmarshalJSON unmarshals the JSON using a fault-tolerant parser that allows comments
// and trailing commas. If any unrecoverable faults are found, an error is returned.
func UnmarshalJSON(text string, v interface{}) error {
	data, err := parseJSON(text)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, v)
}

func parseJSON(text string) ([]byte, error) {
	data, errs := jsonx.Parse(text, jsonx.ParseOptions{Comments: true, TrailingCommas: true})
	if len(errs) > 0 {
		return data, fmt.Errorf("failed to parse JSON: %v", errs)
	}
	return data, nil
}

// NormalizeJSON converts JSON with comments, trailing commas, and some types of syntax errors into
// standard JSON.
func NormalizeJSON(input string) []byte {
	output, _ := jsonx.Parse(string(input), jsonx.ParseOptions{Comments: true, TrailingCommas: true})
	if len(output) == 0 {
		return []byte("{}")
	}
	return output
}
