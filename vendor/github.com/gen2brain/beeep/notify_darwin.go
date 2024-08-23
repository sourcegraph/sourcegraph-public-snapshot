// +build darwin,!linux,!freebsd,!netbsd,!openbsd,!windows,!js

package beeep

import (
	"os/exec"
)

// Notify sends desktop notification.
//
// On macOS this executes AppleScript with `osascript` binary.
func Notify(title, message, appIcon string) error {
	osa, err := exec.LookPath("osascript")
	if err != nil {
		return err
	}

	cmd := exec.Command(osa, "-e", `display notification "`+message+`" with title "`+title+`"`)
	return cmd.Run()
}
