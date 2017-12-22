package env

import (
	"fmt"
	"log"
	"os"
	"sort"
)

var descriptions = make(map[string]string)
var locked = false

var Version = Get("VERSION", "dev", "the version of the packaged app, usually set by Dockerfile")

// Get returns the value of the given environment variable. It also registers the description for
// PrintHelp. Calling Get with the same name twice causes a panic. Get should only be called on
// package initialization. Calls at a later point will cause a panic if Lock was called before.
//
// This should be used for only *internal* environment values. User-visible configuration should be
// added to the Config struct in the sourcegraph.com/sourcegraph/sourcegraph/config package.
func Get(name string, defaultValue string, description string) string {
	if locked {
		panic("env.Get has to be called on package initialization")
	}

	if _, ok := descriptions[name]; ok {
		panic(fmt.Sprintf("%q already registered", name))
	}

	if defaultValue != "" {
		description = fmt.Sprintf("%s (default: %q)", description, defaultValue)
	}
	descriptions[name] = description

	value, ok := os.LookupEnv(name)
	if !ok {
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
