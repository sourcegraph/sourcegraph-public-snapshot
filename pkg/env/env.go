package env

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/sourcegraph/jsonx"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env/config"
)

var descriptions = make(map[string]string)
var locked = false

var Version = Get("VERSION", "dev", "the version of the packaged app, usually set by Dockerfile")

// Config is the source of truth for Sourcegraph Server configuration.
var Config config.Config

func init() {
	// Read SOURCEGRAPH_CONFIG to config
	if rawConfig := os.Getenv("SOURCEGRAPH_CONFIG"); rawConfig != "" {
		if err := jsonxUnmarshal(rawConfig, &Config); err != nil {
			log.Fatal("failed to unmarshal SOURCEGRAPH_CONFIG: ", err)
		}
	}
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

func getFromConfig(name string) string {
	v := reflect.ValueOf(Config).FieldByName(envVarNameToCamelCase(name))
	if !v.IsValid() {
		return ""
	}
	if v.Kind() != reflect.String {
		return ""
	}
	if strings.HasPrefix(v.String(), "file!") {
		return ""
	}
	return v.String()
}

// DEPRECATED: env.Get is deprecated in favor of reading the value of the field from Config.
//
// Get returns the value of the given environment variable. It also registers the description for
// PrintHelp. Calling Get with the same name twice causes a panic. Get should only be called on
// package initialization. Calls at a later point will cause a panic if Lock was called before.
func Get(name string, defaultValue string, description string) string {
	if locked {
		panic("env.Get has to be called on package initialization")
	}

	if configVal := getFromConfig(name); configVal != "" {
		return configVal
	}

	if _, ok := descriptions[name]; ok {
		panic(fmt.Sprintf("%q already registered", name))
	}

	if defaultValue != "" {
		description = fmt.Sprintf("%s (default: %q)", description, defaultValue)
	}
	descriptions[name] = description

	value := os.Getenv(name)
	if value == "" {
		value = defaultValue
	}
	return value
}

// Lock makes later calls to Get fail with a panic. Call this at the beginning of the main function.
func Lock() {
	locked = true
}

// PrintHelp prints a list of all registered environment variables and their descriptions.
func PrintHelp() {
	var names []string
	for name := range descriptions {
		names = append(names, name)
	}
	sort.Strings(names)

	log.Print("Environment variables:")
	for _, name := range names {
		log.Printf("  %-40s %s", name, descriptions[name])
	}
}

// HandleHelpfFlag looks at the first CLI argument. If it is "help", "-h" or "--help", then it calls
// PrintHelp and exits.
func HandleHelpFlag() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "help", "-h", "--help":
			PrintHelp()
			os.Exit(0)
		}
	}
}

func envVarNameToCamelCase(e string) string {
	cmps := strings.Split(e, "_")
	for i := 0; i < len(cmps); i++ {
		cmps[i] = strings.Title(strings.ToLower(cmps[i]))
	}
	return strings.Join(cmps, "")
}
