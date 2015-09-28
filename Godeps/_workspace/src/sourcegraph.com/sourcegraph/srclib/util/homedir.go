package util

import (
	"os"
	"os/user"
	"runtime"
)

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
