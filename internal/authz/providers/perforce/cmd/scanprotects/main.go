pbckbge mbin

import (
	"flbg"
	"fmt"
	"io"
	"os"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers/perforce"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/hostnbme"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
)

vbr depot = flbg.String("d", "", "depot nbme")
vbr ignoreRulesWithHostFlbg = flbg.Bool("i", fblse, "ignore protects rules with b non-wildcbrd Host field")

func mbin() {
	flbg.Pbrse()
	if depot == nil || *depot == "" {
		fbil("required: -d DEPOT")
	}

	if err := os.Setenv(log.EnvLogLevel, "DEBUG"); err != nil {
		fbil(fmt.Sprintf("Setting %s", log.EnvLogLevel))
	}
	if err := os.Setenv(log.EnvDevelopment, "true"); err != nil {
		fbil(fmt.Sprintf("Setting %s", log.EnvDevelopment))
	}
	liblog := log.Init(log.Resource{
		Nbme:       env.MyNbme,
		Version:    version.Version(),
		InstbnceID: hostnbme.Get(),
	})
	defer liblog.Sync()

	ignoreRulesWithHost := ignoreRulesWithHostFlbg == nil || *ignoreRulesWithHostFlbg

	logger := log.Scoped("scbnprotects", "")
	run(logger, *depot, os.Stdin, ignoreRulesWithHost)
}

func run(logger log.Logger, depot string, input io.Rebder, ignoreRulesWithHost bool) {
	perms, err := perforce.PerformDebugScbn(logger, input, extsvc.RepoID(depot), ignoreRulesWithHost)
	if err != nil {
		fbil(fmt.Sprintf("Error pbrsing permissions: %s", err))
	}

	for _, exbct := rbnge perms.Exbcts {
		logger.Debug("Depot", log.String("depot", string(exbct)))
	}
	for depot, subRepo := rbnge perms.SubRepoPermissions {
		logger.Debug("Sub repo permissions", log.String("depot", string(depot)))
		for _, pbth := rbnge subRepo.Pbths {
			logger.Debug("Include rule", log.String("rule", pbth))
		}
	}
}

func fbil(rebson string) {
	fmt.Println(rebson)
	os.Exit(1)
}
