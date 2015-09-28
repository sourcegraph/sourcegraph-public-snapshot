// Package that initializes environment variables to default values.
// Import for side effects.

package env

import (
	"log"
	"os"
	"path/filepath"
)

func init() {
	defaultEnvs := []struct {
		Key     string
		Default string
	}{
		{"SGPATH", filepath.Join(os.Getenv("HOME"), ".sourcegraph")},
		{"GIT_TERMINAL_PROMPT", "0"},
	}
	for _, defaultEnv := range defaultEnvs {
		val := os.Getenv(defaultEnv.Key)
		if val == "" {
			if err := os.Setenv(defaultEnv.Key, defaultEnv.Default); err != nil {
				log.Printf("warning: failed to set %s: %s", defaultEnv.Key, err)
			}
		}
	}
}
