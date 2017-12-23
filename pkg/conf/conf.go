package conf

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/sourcegraph/jsonx"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

var configFilePath = os.Getenv("SOURCEGRAPH_CONFIG_FILE")

var raw = func() string {
	v, ok := os.LookupEnv("SOURCEGRAPH_CONFIG")
	if ok {
		if configFilePath != "" {
			log.Fatal("Multiple configuration sources are not allowed. Use only one of SOURCEGRAPH_CONFIG and SOURCEGRAPH_CONFIG_FILE env vars.")
		}
		return v
	}
	if configFilePath == "" {
		return ""
	}
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("Error reading configuration file %s: %s", configFilePath, err)
	}
	return string(data)
}()

// Raw returns the raw site configuration JSON. Its value is constant for the entire
// lifetime of this process.
func Raw() string { return raw }

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
	if raw != "" {
		if err := jsonxUnmarshal(raw, &cfg); err != nil {
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

// FilePath is the path to the configuration file, if any. The actual data is
// always read from SOURCEGRAPH_CONFIG, not from this file, to avoid race conditions
// (reloading config without a process restart is not yet supported).
func FilePath() string { return configFilePath }

// Write writes the JSON configuration to the config file. If the file is unknown
// or it's not editable, an error is returned. Currently restartToApply is always
// true, to indicate that the server must be restarted to apply the updated config.
func Write(input string) (restartToApply bool, err error) {
	if !IsWritable() {
		return false, errors.New("configuration is not writable")
	}

	if err := ioutil.WriteFile(configFilePath, []byte(input), 0600); err != nil {
		return false, err
	}
	return true, nil
}

// IsWritable reports whether the config can be overwritten.
func IsWritable() bool { return configFilePath != "" }

// IsDirty reports whether the config has been changed since this process started.
// This can occur when config is read from a file and the file has changed on disk.
// Reloading config (without a process restart) is not currently supported.
func IsDirty() bool {
	if configFilePath == "" {
		return false // env var config can't change
	}
	data, err := ioutil.ReadFile(configFilePath)
	return err != nil || string(data) != raw
}
