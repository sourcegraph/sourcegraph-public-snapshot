package confvalidators

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
)

func validateExternalURL(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	if deploy.IsDeployTypeSingleDockerContainer(deploy.Type()) || deploy.IsApp() {
		return nil
	}

	if c.SiteConfig().ExternalURL == "" {
		problems = append(problems, conf.NewSiteProblem("`externalURL` is required to be set for many features of Sourcegraph to work correctly."))
	}

	return problems
}
