package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/jsonx"
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

	// fileWrite signals when our app writes to the configuration file. The
	// secondary channel is closed when conf.Get() would return the new
	// configuration that has been written to disk.
	fileWrite = make(chan chan struct{}, 1)
)

func init() {
	// Read configuration initially.
	if err := initConfig(false); err != nil {
		log.Fatalf("failed to read configuration from environment: %s. Fix your Sourcegraph configuration (%s) to resolve this error. Visit https://about.sourcegraph.com/docs to learn more.", err, configFilePath)
	}

	// Every five seconds, check if the configuration has changed and notify
	// watchers when it has.
	go func() {
		for {
			var signalDoneReading chan struct{}
			select {
			case signalDoneReading = <-fileWrite:
				// File was changed on FS, so check now.
			case <-time.After(5 * time.Second):
				// File possibly changed on FS, so check now.
			}
			if IsDirty() {
				// Read the new configuration from disk.
				if err := initConfig(true); err != nil {
					log.Printf("failed to read configuration from environment: %s. Fix your Sourcegraph configuration (%s) to resolve this error. Visit https://about.sourcegraph.com/docs to learn more.", err, configFilePath)
				}
				if signalDoneReading != nil {
					close(signalDoneReading)
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
			} else if signalDoneReading != nil {
				close(signalDoneReading)
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

func parseConfig(data string) (*schema.SiteConfiguration, error) {
	var tmpConfig schema.SiteConfiguration

	// SOURCEGRAPH_CONFIG takes lowest precedence.
	if data != "" {
		if err := UnmarshalJSON(data, &tmpConfig); err != nil {
			return nil, err
		}
	}

	// Env var config takes highest precedence but is deprecated.
	if v, envVarNames, err := configFromEnv(); err != nil {
		return nil, err
	} else if len(envVarNames) > 0 {
		if err := json.Unmarshal(v, &tmpConfig); err != nil {
			return nil, err
		}
	}

	// For convenience, make sure this is not nil.
	if tmpConfig.ExperimentalFeatures == nil {
		tmpConfig.ExperimentalFeatures = &schema.ExperimentalFeatures{}
	}
	return &tmpConfig, nil
}

func initConfig(reinitialize bool) error {
	rawConfig, err := readConfig()
	if err != nil {
		return err
	}

	rawMu.Lock()
	raw = rawConfig
	rawMu.Unlock()

	tmpConfig, err := parseConfig(rawConfig)
	if err != nil {
		return err
	}

	cfgMu.Lock()
	defer cfgMu.Unlock()
	if reinitialize {
		// Update global "needs restart" state.
		if needRestartToApply(cfg, tmpConfig) {
			markNeedServerRestart()
		}
	}
	cfg = tmpConfig
	return nil
}

// FilePath is the path to the configuration file, if any.
func FilePath() string { return configFilePath }

// requireRestart is a list of configuration properties that require restarting
// the given services to take effect.
//
// TODO(slimsag): make use of this list.
var requireRestart = map[string][]string{
	// These properties initialize global tracers that cannot be swapped out at
	// runtime. They affect basically all services.
	//
	// TODO(slimsag): We could change this at runtime if we made a patch to
	// opentracing that made SetGlobalTracer atomic, but it may be hard to get
	// that accepted.
	"lightstepAccessToken": []string{"all"},
	"lightstepProject":     []string{"all"},
	"useJaeger":            []string{"all"},
}

// doNotRequireRestart is a list of options that do not require a service restart.
//
// TODO(slimsag): eliminate the need for this once all conf.GetTODO are removed.
var doNotRequireRestart = []string{
	"github",
	"gitlab",
	"phabricator",
	"bitbucketServer",
	"repos.list",
	"gitMaxConcurrentClones",
	"repoListUpdateInterval",
	"gitolite",
	"gitOriginMap",
	"githubClientID",
	"githubClientSecret",
	"settings",
	"secretKey",
	"htmlHeadTop",
	"htmlHeadBottom",
	"htmlBodyTop",
	"htmlBodyBottom",
	"disableBuiltInSearches",
	"disableExampleSearches",
	"email.smtp",
	"email.address",
	"disableAutoGitUpdates",
	"corsOrigin",
	"dontIncludeSymbolResultsByDefault",
	"langservers",
}

// Write writes the JSON configuration to the config file. If the file is unknown
// or it's not editable, an error is returned.
func Write(input string) error {
	if !IsWritable() {
		return errors.New("configuration is not writable")
	}

	// Parse the configuration so that we can diff it (this also validates it
	// is proper JSON).
	after, err := parseConfig(input)
	if err != nil {
		return err
	}

	before := Get()

	if err := ioutil.WriteFile(configFilePath, []byte(input), 0600); err != nil {
		return err
	}

	// Wait for the change to the configuration file to be detected. Otherwise
	// we would return to the caller earlier than conf.Get() would return the
	// new configuration.
	doneReading := make(chan struct{}, 1)
	fileWrite <- doneReading
	<-doneReading

	// Update global "needs restart" state.
	if needRestartToApply(before, after) {
		markNeedServerRestart()
	}
	return nil
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

// needRestartToApply determines if a restart is needed to apply the changes
// between the two configurations.
func needRestartToApply(before, after *schema.SiteConfiguration) bool {
	diff := diff(before, after)

	// Delete fields that do not require a process restart from the diff. Then
	// len(diff) > 0 tells us if we need to restart or not.
	for _, option := range doNotRequireRestart {
		delete(diff, option)
	}
	return len(diff) > 0
}

// diff returns names of the Go fields that have different values between the
// two configurations.
func diff(before, after *schema.SiteConfiguration) (fields map[string]struct{}) {
	fields = make(map[string]struct{})
	b := reflect.ValueOf(before).Elem()
	a := reflect.ValueOf(after).Elem()
	for i := 0; i < b.NumField(); i++ {
		beforeField := b.Field(i)
		afterField := a.Field(i)

		tag := b.Type().Field(i).Tag.Get("json")
		if tag == "" {
			// should never happen, and if it does this diffing func cannot work.
			panic(fmt.Sprintf("missing json struct field tag on schema.SiteConfiguration field %q", b.Type().Field(i).Name))
		}
		if !reflect.DeepEqual(beforeField.Interface(), afterField.Interface()) {
			fieldName := parseJSONTag(tag)
			fields[fieldName] = struct{}{}
		}
	}
	return fields
}

// parseJSONTag parses a JSON struct field tag to return the JSON field name.
func parseJSONTag(tag string) string {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx]
	}
	return tag
}

// Edit invokes the provided function to compute edits to the site
// configuration. It then applies and writes them.
//
// The computation function is provided the current configuration, which should
// NEVER be modified in any way. Always copy values.
func Edit(computeEdits func(current *schema.SiteConfiguration, raw string) ([]jsonx.Edit, error)) error {
	current := Get()
	raw := Raw()

	// Compute edits.
	edits, err := computeEdits(current, raw)
	if err != nil {
		return errors.Wrap(err, "computeEdits")
	}

	// Apply edits and write out new configuration.
	newConfig, err := jsonx.ApplyEdits(raw, edits...)
	if err != nil {
		return errors.Wrap(err, "jsonx.ApplyEdits")
	}
	err = Write(newConfig)
	if err != nil {
		return errors.Wrap(err, "conf.Write")
	}
	return nil
}

// FormatOptions is the default format options that should be used for jsonx
// edit computation.
var FormatOptions = jsonx.FormatOptions{InsertSpaces: true, TabSize: 2, EOL: "\n"}

var (
	needRestartMu sync.RWMutex
	needRestart   bool
)

// NeedServerRestart tells if the server needs to restart for pending configuration
// changes to take effect.
func NeedServerRestart() bool {
	needRestartMu.RLock()
	defer needRestartMu.RUnlock()
	return needRestart
}

// markNeedServerRestart marks the server as needing a restart so that pending
// configuration changes can take effect.
func markNeedServerRestart() {
	needRestartMu.Lock()
	needRestart = true
	needRestartMu.Unlock()
}
