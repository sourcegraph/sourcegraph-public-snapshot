package conf

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"time"

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

// Raw returns the raw site configuration JSON.
func Raw() string { return raw }

// Get returns a copy of the configuration. The returned value should NEVER be
// modified.
//
// Important: The configuration can change while the process is running! Code
// should only call this in response to conf.Watch OR it should invoke it
// periodically to ensure it responds to configuration changes while the
// process is running.
//
// There are a select few configuration options that do restart the server (for
// example, TLS or which port the frontend listens on) but these are the
// exception rather than the rule. In general, ANY use of configuration should
// be done in such a way that it responds to config changes while the process
// is running.
func Get() schema.SiteConfiguration {
	if MockGetData != nil {
		return *MockGetData
	}
	return cfg
}

// Watch calls the given function whenever the configuration has changed. The
// new configuration can be recieved by calling conf.Get.
//
// Before Watch returns, it will invoke f to use the current configuration.
func Watch(f func()) {
	// Call the function now, to use the current configuration.
	f()

	// Every five seconds, check if the configuration has changed and invoke f.
	for {
		time.Sleep(5 * time.Second)
		if IsDirty() {
			f()
		}
	}
}

// MockGetData is overridden in tests that need to mock site config.
var MockGetData *schema.SiteConfiguration

// cfg is initialized to configuration defaults.
var cfg = schema.SiteConfiguration{
	MaxReposToSearch: 500,
}

func init() {
	// Read env vars to config
	if err := initConfig(); err != nil {
		log.Fatalf("failed to read configuration from environment: %s. Fix your Sourcegraph configuration (%s) to resolve this error. Visit https://about.sourcegraph.com/docs to learn more.", err, configFilePath)
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
	if v, envVarNames, err := configFromEnv(); err != nil {
		return err
	} else if len(envVarNames) > 0 {
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

// FilePath is the path to the configuration file, if any.
func FilePath() string { return configFilePath }

// Write writes the JSON configuration to the config file. If the file is unknown
// or it's not editable, an error is returned. restartToApply indicates whether
// or not the server must be restarted to apply the updated config.
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
func IsDirty() bool {
	if configFilePath == "" {
		return false // env var config can't change
	}
	data, err := ioutil.ReadFile(configFilePath)
	return err != nil || string(data) != raw
}
