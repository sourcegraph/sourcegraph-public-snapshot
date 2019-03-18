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
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	configFilePath string

	rawMu sync.RWMutex
	raw   string
)

// Raw returns the raw site configuration JSON.
func Raw() string {
	rawMu.RLock()
	defer rawMu.RUnlock()
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
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if mockGetData != nil {
		return mockGetData
	}
	return cfg
}

var mockGetData *schema.SiteConfiguration

// Mock sets up mock data for the site configuration. It uses the configuration
// mutex, to avoid possible races between test code and possible config watchers.
func Mock(mockery *schema.SiteConfiguration) {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	mockGetData = mockery
}

// GetTODO denotes code that may or may not be using configuration correctly.
// The code may need to be updated to use conf.Watch, or it may already be e.g.
// invoked only in response to a user action (in which case it does not need to
// use conf.Watch). See Get documentation for more details.
func GetTODO() *schema.SiteConfiguration {
	return Get()
}

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
	configFilePath = os.Getenv("SOURCEGRAPH_CONFIG_FILE")

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
	if configFilePath == "" {
		return "", nil
	}
	data, err := ioutil.ReadFile(configFilePath)
	return string(data), err
}

// ParseConfigData reads the provided config string, but NOT the environment
func ParseConfigData(data string) (*schema.SiteConfiguration, error) {
	var tmpConfig schema.SiteConfiguration

	if data != "" {
		data, err := jsonc.Parse(data)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &tmpConfig); err != nil {
			return nil, err
		}
	}

	// For convenience, make sure this is not nil.
	if tmpConfig.ExperimentalFeatures == nil {
		tmpConfig.ExperimentalFeatures = &schema.ExperimentalFeatures{}
	}
	return &tmpConfig, nil
}

// parseConfig reads the provided string, then merges in additional
// data from the (deprecated) environment.
func parseConfig(data string) (*schema.SiteConfiguration, error) {
	tmpConfig, err := ParseConfigData(data)
	if err != nil {
		return nil, err
	}

	// Env var config takes highest precedence but is deprecated.
	if v, envVarNames, err := configFromEnv(); err != nil {
		return nil, err
	} else if len(envVarNames) > 0 {
		if err := json.Unmarshal(v, tmpConfig); err != nil {
			return nil, err
		}
	}
	return tmpConfig, nil
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

// TODO(slimsag): add back requireRestart and make use of it (it is a list of config properties that
// require restarting the given services to take effect)

// doNotRequireRestart is a list of options that do not require a service restart.
//
// TODO(slimsag): eliminate the need for this once all conf.GetTODO are removed.
var doNotRequireRestart = []string{
	"auth.allowSignup",
	"auth.public",
	"auth.userIdentityHTTPHeader",
	"github",
	"gitlab",
	"phabricator",
	"awsCodeCommit",
	"bitbucketServer",
	"repos.list",
	"gitMaxConcurrentClones",
	"repoListUpdateInterval",
	"gitolite",
	"gitOriginMap",
	"githubClientID",
	"githubClientSecret",
	"settings",
	"htmlHeadTop",
	"htmlHeadBottom",
	"htmlBodyTop",
	"htmlBodyBottom",
	"httpStrictTransportSecurity",
	"httpToHttpsRedirect",
	"disableBuiltInSearches",
	"email.smtp",
	"email.address",
	"disableAutoGitUpdates",
	"corsOrigin",
	"dontIncludeSymbolResultsByDefault",
	"langservers",
	"platform",
	"log",
	"experimentalFeatures::jumpToDefOSSIndex",
	"experimentalFeatures::canonicalURLRedirect",
	"experimentalFeatures::multipleAuthProviders",
	"experimentalFeatures::platform",
	"experimentalFeatures::discussions",
	"reviewBoard",
	"parentSourcegraph",
	"maxReposToSearch",
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

// merge a map, overwriting keys
func mergeMap(destMap, srcMap reflect.Value) {
	mapType := destMap.Type()
	if mapType.Kind() != reflect.Map {
		fmt.Printf("error: not a map: %T\n", destMap)
		return
	}
	valueType := mapType.Elem()
	zero := reflect.Zero(valueType)
	keys := srcMap.MapKeys()
	for _, key := range keys {
		srcValue := srcMap.MapIndex(key)
		destValue := destMap.MapIndex(key)
		switch srcValue.Kind() {
		case reflect.Struct:
			if destValue.IsNil() {
				destMap.SetMapIndex(key, srcValue)
			} else {
				mergeStruct(destValue.Interface(), srcValue.Interface())
			}
		case reflect.Slice:
			destMap.SetMapIndex(key, reflect.AppendSlice(destValue, srcValue))
		case reflect.Map:
			mergeMap(destValue, srcValue)
		default:
			if srcValue.Interface() != zero.Interface() {
				destMap.SetMapIndex(key, srcValue)
			}
		}
		destMap.SetMapIndex(key, srcMap.MapIndex(key))
	}
}

// merge a struct. recurse on structs, append arrays,
// overwrite everything else.
func mergeStruct(destInterface, srcInterface interface{}) {
	destType := reflect.TypeOf(destInterface)
	dest := reflect.ValueOf(destInterface)
	if destType.Kind() == reflect.Ptr {
		dest = dest.Elem()
		destType = dest.Type()
	}
	srcType := reflect.TypeOf(srcInterface)
	src := reflect.ValueOf(srcInterface)
	if srcType.Kind() == reflect.Ptr {
		src = src.Elem()
		srcType = src.Type()
	}
	if destType != srcType {
		fmt.Printf("fatal: destType '%T' and srcType '%T' are not equal.\n", dest, src)
		return
	}
	for i := 0; i < destType.NumField(); i++ {
		destField := dest.Field(i)
		srcField := src.Field(i)
		zero := reflect.Zero(destField.Type())
		switch destField.Kind() {
		case reflect.Struct:
			mergeStruct(destField, srcField)
		case reflect.Slice:
			destField.Set(reflect.AppendSlice(destField, srcField))
		case reflect.Map:
			mergeMap(destField, srcField)
		case reflect.Ptr:
			switch destField.Elem().Kind() {
			case reflect.Struct:
				srcValid := srcField.Elem().IsValid()
				destValid := destField.Elem().IsValid()
				if srcValid {
					if destValid {
						mergeStruct(destField.Interface(), srcField.Interface())
					} else {
						destField.Set(srcField)
					}
				}
			case reflect.Slice:
				destField.Elem().Set(reflect.AppendSlice(destField.Elem(), srcField.Elem()))
			case reflect.Map:
				mergeMap(destField.Elem(), srcField.Elem())
			}
		default:
			if srcField.Interface() != zero.Interface() {
				destField.Set(srcField)
			}
		}
	}
}

// recursively merge components of site config
func AppendConfig(dest, src *schema.SiteConfiguration) *schema.SiteConfiguration {
	if dest == nil {
		return src
	}
	if src == nil {
		return dest
	}
	mergeStruct(dest, src)
	return dest
}

// IsWritable reports whether the config can be overwritten.
func IsWritable() bool { return configFilePath != "" }

// IsDirty reports whether the config has been changed since this process started.
// This can occur when config is read from a file and the file has changed on disk.
func IsDirty() bool {
	if configFilePath == "" {
		return false
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
	beforeFields := getJSONFields(before)
	afterFields := getJSONFields(after)
	for fieldName, beforeField := range beforeFields {
		afterField := afterFields[fieldName]
		if !reflect.DeepEqual(beforeField, afterField) {
			fields[fieldName] = struct{}{}
		}
	}
	return fields
}

func getJSONFields(vv interface{}) (fields map[string]interface{}) {
	fields = make(map[string]interface{})
	v := reflect.ValueOf(vv).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		tag := v.Type().Field(i).Tag.Get("json")
		if tag == "" {
			// should never happen, and if it does this func cannot work.
			panic(fmt.Sprintf("missing json struct field tag on %T field %q", v.Interface(), v.Type().Field(i).Name))
		}
		if ef, ok := f.Interface().(*schema.ExperimentalFeatures); ok && ef != nil {
			for fieldName, fieldValue := range getJSONFields(ef) {
				fields["experimentalFeatures::"+fieldName] = fieldValue
			}
			continue
		}
		fieldName := parseJSONTag(tag)
		fields[fieldName] = f.Interface()
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
