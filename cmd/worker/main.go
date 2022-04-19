package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

func main() {
	log.Init(log.Resource{
		Name:    "worker",
		Version: version.Version(),
	})

	authz.SetProviders(true, []authz.Provider{})
	shared.Start(nil, nil)
}
