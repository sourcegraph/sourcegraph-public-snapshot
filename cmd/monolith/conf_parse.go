package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"os"
	"unicode"

	"github.com/sourcegraph/jsonx"
)

// setDefaultEnv will set the environment variable if it is not set.
func setDefaultEnv(k, v string) string {
	if s, ok := os.LookupEnv(k); ok {
		return s
	}
	err := os.Setenv(k, v)
	if err != nil {
		log.Fatal(err)
	}
	return v
}

// setDefaultEnvFromConfig parses a json value and updates the environ with
// the fields present.
func setDefaultEnvFromConfig(rawConfig string) {
	env, err := getEnvironFromConfig(rawConfig)
	log.Fatal("failed to unmarshal SOURCEGRAPH_CONFIG: ", err)
	for k, v := range env {
		setDefaultEnv(k, v)
	}
}

// getEnvironFromConfig parses a json value into an environ map.
func getEnvironFromConfig(rawConfig string) (map[string]string, error) {
	raw := map[string]*json.RawMessage{}
	if err := jsonxUnmarshal(rawConfig, &raw); err != nil {
		return nil, err
	}
	environ := map[string]string{}
	for k, v := range raw {
		// Convert key from camelCase into CAMEL_CASE
		var b bytes.Buffer
		inword := false
		for _, c := range k {
			if inword && unicode.IsUpper(c) {
				b.WriteRune('_')
			}
			inword = !unicode.IsUpper(c)
			b.WriteRune(unicode.ToUpper(c))
		}
		k = b.String()

		// Convert into an environ value
		s := string(*v)
		// The only value we "unwrap" is a string
		if len(s) > 0 && s[0] == '"' {
			if err := json.Unmarshal([]byte(*v), &s); err != nil {
				// This shouldn't happen
				return nil, err
			}
		}

		environ[k] = s
	}
	return environ, nil
}

// jsonxUnmarshal unmarshals the JSON using a fault tolerant parser. If any
// unrecoverable faults are found an error is returned
func jsonxUnmarshal(text string, v interface{}) error {
	data, errs := jsonx.Parse(text, jsonx.ParseOptions{Comments: true, TrailingCommas: true})
	if len(errs) > 0 {
		return errors.New("failed to parse json")
	}
	return json.Unmarshal(data, v)
}
