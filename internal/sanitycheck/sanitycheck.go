package sanitycheck

import (
	"fmt"
	"os"
)

// Pass exits with a code zero if the environment variable SANITY_CHECK equals
// to "true". This enables testing that the current program is in a runnable
// state against the platform it's being executed on.
//
// See https://github.com/GoogleContainerTools/container-structure-test
func Pass() {
	if os.Getenv("SANITY_CHECK") == "true" {
		fmt.Println("Sanity check passed, exiting without error")
		os.Exit(0)
	}
}
