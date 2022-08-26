package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/authz/perforce"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

var depot = flag.String("d", "", "depot name")

func main() {
	flag.Parse()
	if depot == nil || *depot == "" {
		fail("required: -d DEPOT")
	}

	if err := os.Setenv(log.EnvLogLevel, "DEBUG"); err != nil {
		fail("Setting SRC_LOG_LEVEL")
	}
	if err := os.Setenv(log.EnvDevelopment, "true"); err != nil {
		fail("Setting SRC_LOG_LEVEL")
	}
	liblog := log.Init(log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	})
	defer liblog.Sync()

	logger := log.Scoped("scanprotects", "")
	run(logger, *depot, os.Stdin)
}

func run(logger log.Logger, depot string, input io.Reader) {
	perms, err := perforce.PerformDebugScan(logger, input, extsvc.RepoID(depot))
	if err != nil {
		fail(fmt.Sprintf("Error parsing permissions: %s", err))
	}

	for _, exact := range perms.Exacts {
		logger.Debug("Depot", log.String("depot", string(exact)))
	}
	for depot, subRepo := range perms.SubRepoPermissions {
		logger.Debug("Sub repo permissions", log.String("depot", string(depot)))
		for _, include := range subRepo.PathIncludes {
			logger.Debug("Include rule", log.String("rule", include))
		}
		for _, exclude := range subRepo.PathExcludes {
			logger.Debug("Include rule", log.String("rule", exclude))
		}
	}
}

func fail(reason string) {
	fmt.Println(reason)
	os.Exit(1)
}
