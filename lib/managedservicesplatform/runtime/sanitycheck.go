package runtime

import (
	"fmt"
	"os"
)

// passSanityCheck exits with a code zero if the environment variable SANITY_CHECK equals
// to "true". See internal/sanitycheck.
func passSanityCheck() {
	if os.Getenv("SANITY_CHECK") == "true" {
		fmt.Println("Sanity check passed, exiting without error")
		os.Exit(0)
	}
}
