package conf

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (
	BuildLogDir = GetenvOrDefault("SG_BUILD_LOG_DIR", filepath.Join(TempDir(), "sg-log/build"))
)

func MustGetenv(name string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	log.Fatalf("Fatal error: Environment variable %s must be set.", name)
	panic("unreachable")
}

// GetenvBool parses the env var with the given name as a boolean
// (using strconv.ParseBool) and returns the boolean value. If boolean
// parsing fails (e.g., the env var is empty), it returns false.
func GetenvBool(name string) bool {
	v, _ := strconv.ParseBool(os.Getenv(name))
	return v
}

func GetenvOrDefault(name string, defaultValue string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return defaultValue
}

func GetenvURLOrDefault(name string, defaultURLStr string) *url.URL {
	urlStr := GetenvOrDefault(name, defaultURLStr)
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Fatalf("Fatal error: Environment variable %s must contain a valid URL, not %q.", name, urlStr)
	}
	return u
}

func GetenvURL(name string) *url.URL {
	urlStr := os.Getenv(name)
	if urlStr == "" {
		return nil
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Fatalf("Fatal error: Environment variable %s must contain a valid URL, not %q.", name, urlStr)
	}
	return u
}

func GetenvDurationOrDefault(name string, defaultDurationStr string) time.Duration {
	durationStr := GetenvOrDefault(name, defaultDurationStr)
	d, err := time.ParseDuration(durationStr)
	if err != nil {
		log.Fatalf("Fatal error: Environment variable %s must contain a valid duration string, not %q.", name, durationStr)
	}
	return d
}

func GetenvIntOrDefault(name string, defaultVal int) int {
	s := os.Getenv(name)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("Fatal error: Environment variable %s must contain a valid number (or be empty), not %q.", name, s)
	}
	return v
}

func GetenvInt(name string) int {
	s := os.Getenv(name)
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("Fatal error: Environment variable %s must contain a valid number (or be empty), not %q.", name, s)
	}
	return v
}
