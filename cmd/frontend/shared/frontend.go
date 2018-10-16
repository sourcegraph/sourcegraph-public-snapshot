// Package shared contains the frontend command implementation shared
//
package shared

import (
	"fmt"
	"log"
	"os"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/assets"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli"
	"github.com/sourcegraph/sourcegraph/pkg/env"

	// Register Sourcegraph documentation for /help.
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/docsite"
)

// Main is the main function that runs the frontend process.
//
// It is exposed as function in a package so that it can be called by other
// main package implementations such as Sourcegraph Enterprise, which import
// proprietary/private code.
func Main() {
	AssertRequired()
	env.Lock()
	err := cli.Main()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

// AssertRequired fails if certain necessary variables have not been set.
func AssertRequired() {
	if assets.Assets == nil {
		log.Fatal("github.com/sourcegraph/sourcegraph/cmd/frontend/assets.Assets must be non-nil.")
	}
}
