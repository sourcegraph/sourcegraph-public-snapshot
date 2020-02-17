// Package shared contains the frontend command implementation shared
package shared

import (
	"fmt"
	"net/http"
	"os"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli"
	"github.com/sourcegraph/sourcegraph/internal/env"

	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/registry"
)

// Main is the main function that runs the frontend process.
//
// It is exposed as function in a package so that it can be called by other
// main package implementations such as Sourcegraph Enterprise, which import
// proprietary/private code.
func Main(githubWebhook, bitbucketServerWebhook http.Handler) {
	env.Lock()
	err := cli.Main(githubWebhook, bitbucketServerWebhook)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}
