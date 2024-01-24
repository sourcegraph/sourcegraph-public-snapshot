package shared

import (
	"log" //nolint:logging // TODO move all logging to sourcegraph/log
	"os"
)

// SetDefaultEnv will set the environment variable if it is not set.
func SetDefaultEnv(k, v string) string {
	if s, ok := os.LookupEnv(k); ok {
		return s
	}
	err := os.Setenv(k, v)
	if err != nil {
		log.Fatal(err)
	}
	return v
}
