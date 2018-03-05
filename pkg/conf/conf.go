package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

var (
	configFilePath = os.Getenv("SOURCEGRAPH_CONFIG_FILE")

	rawMu sync.RWMutex
	raw   string
)

// Raw returns the raw site configuration JSON.
func Raw() string {
	return raw
}

// Get returns a copy of the configuration. The returned value should NEVER be
// modified.
//
// Important: The configuration can change while the process is running! Code
// should only call this in response to conf.Watch OR it should invoke it
// periodically or in direct response to a user action (e.g. inside an HTTP
// handler) to ensure it responds to configuration changes while the process
// is running.
//
// There are a select few configuration options that do restart the server (for
// example, TLS or which port the frontend listens on) but these are the
// exception rather than the rule. In general, ANY use of configuration should
// be done in such a way that it responds to config changes while the process
// is running.
func Get() *schema.SiteConfiguration {
	if MockGetData != nil {
		return MockGetData
	}
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	return cfg
}

// GetTODO denotes code that may or may not be using configuration correctly.
// The code may need to be updated to use conf.Watch, or it may already be e.g.
// invoked only in response to a user action (in which case it does not need to
// use conf.Watch). See Get documentation for more details.
func GetTODO() *schema.SiteConfiguration {
	return Get()
}

// MockGetData is overridden in tests that need to mock site config.
var MockGetData *schema.SiteConfiguration

var (
	watchersMu sync.Mutex
	watchers   []chan struct{}
)

// Watch calls the given function in a separate goroutine whenever the
// configuration has changed. The new configuration can be received by calling
// conf.Get.
//
// Before Watch returns, it will invoke f to use the current configuration.
func Watch(f func()) {
	// Add the watcher channel now, rather than after invoking f below, in case
	// an update were to happen while we were invoking f.
	notify := make(chan struct{}, 1)
	watchersMu.Lock()
	watchers = append(watchers, notify)
	watchersMu.Unlock()

	// Call the function now, to use the current configuration.
	f()

	go func() {
		// Invoke f when the configuration has changed.
		for {
			<-notify
			f()
		}
	}()
}

var (
	cfgMu sync.RWMutex
	cfg   *schema.SiteConfiguration
)

func init() {
	// Read configuration initially.
	if err := initConfig(); err != nil {
		log.Fatalf("failed to read configuration from environment: %s. Fix your Sourcegraph configuration (%s) to resolve this error. Visit https://about.sourcegraph.com/docs to learn more.", err, configFilePath)
	}

	// Every five seconds, check if the configuration has changed and notify
	// watchers when it has.
	go func() {
		for {
			time.Sleep(5 * time.Second)
			if IsDirty() {
				// Read the new configuration from disk.
				if err := initConfig(); err != nil {
					log.Printf("failed to read configuration from environment: %s. Fix your Sourcegraph configuration (%s) to resolve this error. Visit https://about.sourcegraph.com/docs to learn more.", err, configFilePath)
				}

				watchersMu.Lock()
				for _, watcher := range watchers {
					// Perform a non-blocking send.
					//
					// Since the watcher channels that we are sending on have a
					// buffer of 1, it is guaranteed the watcher will
					// reconsider the config at some point in the future even
					// if this send fails.
					select {
					case watcher <- struct{}{}:
					default:
					}
				}
				watchersMu.Unlock()
			}
		}
	}()
}

func readConfig() (string, error) {
	v, ok := os.LookupEnv("SOURCEGRAPH_CONFIG")
	if ok {
		if configFilePath != "" {
			return "", errors.New("Multiple configuration sources are not allowed. Use only one of SOURCEGRAPH_CONFIG and SOURCEGRAPH_CONFIG_FILE env vars.")
		}
		return v, nil
	}
	if configFilePath == "" {
		return "", nil
	}
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", fmt.Errorf("Error reading configuration file %s: %s", configFilePath, err)
	}
	return string(data), nil
}

func initConfig() error {
	rawConfig, err := readConfig()
	if err != nil {
		return err
	}

	rawMu.Lock()
	raw = rawConfig
	rawMu.Unlock()

	// TODO(slimsag): MaxReposToSearch default value should be in our schema, not here?
	tmpConfig := schema.SiteConfiguration{
		MaxReposToSearch: 500,
	}

	// SOURCEGRAPH_CONFIG takes lowest precedence.
	if raw != "" {
		if err := UnmarshalJSON(raw, &tmpConfig); err != nil {
			return err
		}
	}

	// Env var config takes highest precedence but is deprecated.
	if v, envVarNames, err := configFromEnv(); err != nil {
		return err
	} else if len(envVarNames) > 0 {
		if err := json.Unmarshal(v, &tmpConfig); err != nil {
			return err
		}
	}

	cfgMu.Lock()
	cfg = &tmpConfig
	cfgMu.Unlock()
	return nil
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
