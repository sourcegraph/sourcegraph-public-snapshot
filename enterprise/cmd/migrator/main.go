package main

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/migrator/shared"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

func main() {
	liblog := log.Init(log.Resource{
		Name:    env.MyName,
		Version: version.Version(),
	})
	defer liblog.Sync()

	logger := log.Scoped("migrator", "migrator enterprise edition")

	if err := shared.Start(); err != nil {
		logger.Fatal(err.Error())
	}
}
