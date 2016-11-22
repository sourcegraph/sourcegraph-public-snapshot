package env

import (
	"fmt"
	"os"
	"sort"
)

var descriptions = make(map[string]string)
var locked = false

// Get returns the value of the given environment variable. It also registers the description for
// PrintHelp. Calling Get with the same name twice causes a panic. Get should only be called on
// package initialization. Calls at a later point will cause a panic if Lock was called before.
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

	fmt.Println("Environment vairables:")
	for _, name := range names {
		fmt.Printf("  %-30s %s\n", name, descriptions[name])
	}
}
