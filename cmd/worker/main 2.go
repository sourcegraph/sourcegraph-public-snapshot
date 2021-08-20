package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	"github.com/sourcegraph/sourcegraph/internal/authz"
)

func main() {
	authz.SetProviders(true, []authz.Provider{})
	shared.Start(nil)
}
