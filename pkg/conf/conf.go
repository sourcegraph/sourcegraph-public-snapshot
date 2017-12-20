package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/sourcegraph/jsonx"
	"github.com/xeipuuv/gojsonschema"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

// Get returns a copy of the configuration. The returned value should NEVER be modified.
func Get() schema.SiteConfiguration {
	return cfg
}

// cfg is initialized to configuration defaults.
var cfg = schema.SiteConfiguration{
	MaxReposToSearch: 30,
}

func init() {
	// Read env vars to config
	if err := initConfig(); err != nil {
		log.Fatalf("failed to read configuration from environment: %s", err)
	}
}

func initConfig() error {
	// SOURCEGRAPH_CONFIG takes lowest precedence.
	if v := os.Getenv("SOURCEGRAPH_CONFIG"); v != "" {
		if err := jsonxUnmarshal(v, &cfg); err != nil {
			return err
		}
	}

	// Env var config takes highest precedence but is deprecated.
	if v, err := configFromLegacyEnvVars(); err != nil {
		return err
	} else if len(v) > 0 && string(v) != "{}" {
		log.Printf("Deprecation warning: Add the following config to SOURCEGRAPH_CONFIG instead of passing via other env vars: %s", v)
		if err := json.Unmarshal(v, &cfg); err != nil {
			return err
		}
	}

	return nil
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

// normalizeJSON converts JSON with comments, trailing commas, and some types of syntax errors into
// standard JSON.
func normalizeJSON(input string) []byte {
	output, _ := jsonx.Parse(string(input), jsonx.ParseOptions{Comments: true, TrailingCommas: true})
	if len(output) == 0 {
		return []byte("{}")
	}
	return output
}

// Validate validates the site configuration against its JSON schema.
//
// TODO(sqs): it only validates the SOURCEGRAPH_CONFIG value, not the merged
// config from all env vars. This env var is only used in cmd/server, but it
// is passed onto frontend, so frontend can print useful validation messages
// about it.
func Validate() {
	input := os.Getenv("SOURCEGRAPH_CONFIG")
	if input == "" {
		return
	}

	res, err := gojsonschema.Validate(
		gojsonschema.NewStringLoader(schema.SiteSchemaJSON),
		gojsonschema.NewBytesLoader(normalizeJSON(input)),
	)
	if err != nil {
		log.Printf("Warning: Unable to validate Sourcegraph site configuration: %s", err)
		return
	}
	if !res.Valid() {
		fmt.Fprintln(os.Stderr, "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Fprintln(os.Stderr, "⚠️ Warning: Invalid Sourcegraph site configuration:")
		for _, err := range res.Errors() {
			fmt.Fprintf(os.Stderr, " - %s\n", err.String())
		}
		fmt.Fprintln(os.Stderr, "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}
}
