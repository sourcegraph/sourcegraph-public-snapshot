package env

import (
	"expvar"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

type envflag struct {
	description string
	value       string
}

var (
	env     map[string]envflag
	environ map[string]string
	locked  = false

	expvarPublish = true
)

var (
	// MyName represents the name of the current process.
	MyName, envVarName = findName()
	LogLevel           = Get("SRC_LOG_LEVEL", "warn", "upper log level to restrict log output to (dbug, info, warn, error, crit)")
	LogFormat          = Get("SRC_LOG_FORMAT", "logfmt", "log format (logfmt, condensed, json)")
	LogSourceLink, _   = strconv.ParseBool(Get("SRC_LOG_SOURCE_LINK", "false", "Print an iTerm link to the file:line in VS Code"))
	InsecureDev, _     = strconv.ParseBool(Get("INSECURE_DEV", "false", "Running in insecure dev (local laptop) mode"))
)

// findName returns the name of the current process, that being the
// part of argv[0] after the last slash if any, and also the lowercase
// letters from that, suitable for use as a likely key for lookups
// in things like shell environment variables which can't contain
// hyphens.
func findName() (string, string) {
	// Environment variable names can't contain, for instance, hyphens.
	origName := filepath.Base(os.Args[0])
	name := strings.ReplaceAll(origName, "-", "_")
	if name == "" {
		name = "unknown"
	}
	return origName, name
}

// Ensure behaves like Get except that it sets the environment variable if it doesn't exist.
func Ensure(name, defaultValue, description string) string {
	value := Get(name, defaultValue, description)
	if value == defaultValue {
		err := os.Setenv(name, value)
		if err != nil {
			panic(fmt.Sprintf("failed to set %s environment variable: %v", name, err))
		}
	}

	return value
}

// Get returns the value of the given environment variable. It also registers the description for
// HelpString. Calling Get with the same name twice causes a panic. Get should only be called on
// package initialization. Calls at a later point will cause a panic if Lock was called before.
//
// This should be used for only *internal* environment values.
func Get(name, defaultValue, description string) string {
	if locked {
		panic("env.Get has to be called on package initialization")
	}

	// os.LookupEnv is a syscall. We use Get a lot on startup in many
	// packages. This leads to it being the main contributor to init being
	// slow. So we avoid the constant syscalls by checking env once.
	if environ == nil {
		environ = environMap(os.Environ())
	}

	// Allow per-process override. For instance, SRC_LOG_LEVEL_repo_updater would
	// apply to repo-updater, but not to anything else.
	perProg := name + "_" + envVarName
	value, ok := environ[perProg]
	if !ok {
		value, ok = environ[name]
		if !ok {
			value = defaultValue
		}
	}

	if env == nil {
		env = map[string]envflag{}
	}

	e := envflag{description: description, value: value}
	if existing, ok := env[name]; ok && existing != e {
		panic(fmt.Sprintf("env var %q already registered with a different description or value\n\tBefore: %q\n\tAfter: %q", name, existing, e))
	}
	env[name] = e

	return value
}

// MustGetBytes is similar to Get but ensures that the value is a valid byte size (as defined by go-humanize)
func MustGetBytes(name string, defaultValue string, description string) uint64 {
	s := Get(name, defaultValue, description)
	n, err := humanize.ParseBytes(s)
	if err != nil {
		panic(fmt.Sprintf("parsing environment variable %q. Expected valid time.Duration, got %q", name, s))
	}
	return n
}

// MustGetDuration is similar to Get but ensures that the value is a valid time.Duration.
func MustGetDuration(name string, defaultValue time.Duration, description string) time.Duration {
	s := Get(name, defaultValue.String(), description)
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(fmt.Sprintf("parsing environment variable %q. Expected valid time.Duration, got %q", name, s))
	}
	return d
}

// MustGetInt is similar to Get but ensures that the value is a valid int.
func MustGetInt(name string, defaultValue int, description string) int {
	s := Get(name, strconv.Itoa(defaultValue), description)
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("parsing environment variable %q. Expected valid integer, got %q", name, s))
	}
	return i
}

// MustGetBool is similar to Get but ensures that the value is a valid bool.
func MustGetBool(name string, defaultValue bool, description string) bool {
	s := Get(name, strconv.FormatBool(defaultValue), description)
	b, err := strconv.ParseBool(s)
	if err != nil {
		panic(fmt.Sprintf("parsing environment variable %q. Expected valid bool, got %q", name, s))
	}
	return b
}

func environMap(environ []string) map[string]string {
	m := make(map[string]string, len(environ))
	for _, e := range environ {
		i := strings.Index(e, "=")
		m[e[:i]] = e[i+1:]
	}
	return m
}

// Lock makes later calls to Get fail with a panic. Call this at the beginning of the main function.
func Lock() {
	if locked {
		panic("env.Lock must be called at most once")
	}

	locked = true

	if expvarPublish {
		expvar.Publish("env", expvar.Func(func() any {
			return env
		}))
	}
}

// HelpString prints a list of all registered environment variables and their descriptions.
func HelpString() string {
	helpStr := "Environment variables:\n"

	sorted := make([]string, 0, len(env))
	for name := range env {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)

	for _, name := range sorted {
		e := env[name]
		helpStr += fmt.Sprintf("  %-40s %s (value: %q)\n", name, e.description, e.value)
	}

	return helpStr
}

// HandleHelpFlag looks at the first CLI argument. If it is "help", "-h" or "--help", then it calls
// HelpString and exits.
func HandleHelpFlag() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "help", "-h", "--help":
			fmt.Println(HelpString())
			os.Exit(0)
		}
	}
}
