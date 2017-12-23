package conf

import (
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/sourcegraph/jsonx"
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
	if v, envVarNames, err := configFromLegacyEnvVars(); err != nil {
		return err
	} else if len(envVarNames) > 0 {
		if os.Getenv("DEBUG") != "" {
			log.Printf("Deprecation warning: Add the following config to SOURCEGRAPH_CONFIG instead of passing via other env vars: %v", envVarNames)
		}
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
