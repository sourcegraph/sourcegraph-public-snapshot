package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz/providers/perforce"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

var depot = flag.String("d", "", "depot name")
var ignoreRulesWithHostFlag = flag.Bool("i", false, "ignore protects rules with a non-wildcard Host field")

func main() {
	flag.Parse()
	if depot == nil || *depot == "" {
		fail("required: -d DEPOT")
	}

	if err := os.Setenv(log.EnvLogLevel, "DEBUG"); err != nil {
		fail(fmt.Sprintf("Setting %s", log.EnvLogLevel))
	}
	if err := os.Setenv(log.EnvDevelopment, "true"); err != nil {
		fail(fmt.Sprintf("Setting %s", log.EnvDevelopment))
	}
	liblog := log.Init(log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	})
	defer liblog.Sync()

	ignoreRulesWithHost := ignoreRulesWithHostFlag == nil || *ignoreRulesWithHostFlag

	logger := log.Scoped("scanprotects")
	run(logger, *depot, os.Stdin, ignoreRulesWithHost)
}

func run(logger log.Logger, depot string, input io.Reader, ignoreRulesWithHost bool) {
	perms, err := perforce.PerformDebugScan(logger, input, extsvc.RepoID(depot), ignoreRulesWithHost)
	if err != nil {
		fail(fmt.Sprintf("Error parsing permissions: %s", err))
	}

	for _, exact := range perms.Exacts {
		logger.Debug("Depot", log.String("depot", string(exact)))
	}
	for depot, subRepo := range perms.SubRepoPermissions {
		logger.Debug("Sub repo permissions", log.String("depot", string(depot)))
		for _, path := range subRepo.Paths {
			logger.Debug("Include rule", log.String("rule", path))
		}
	}
}

func fail(reason string) {
	fmt.Println(reason)
	os.Exit(1)
}
