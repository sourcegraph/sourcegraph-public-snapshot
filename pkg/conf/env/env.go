// Package that initializes environment variables to default values.
// Import for side effects.

package env

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
)

var Debug bool

func init() {
	Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
}

// CurrentUserHomeDir tries to get the current user's home directory in a
// cross-platform manner.
func CurrentUserHomeDir() string {
	user, err := user.Current()
	if err == nil && user.HomeDir != "" {
		return user.HomeDir
	}

	// from http://stackoverflow.com/questions/7922270/obtain-users-home-directory
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func init() {
	defaultEnvs := []struct {
		Key     string
		Default string
	}{
		{"SGPATH", filepath.Join(CurrentUserHomeDir(), ".sourcegraph")},
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
