package authz

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/authz/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/authz/github"
	"github.com/sourcegraph/sourcegraph/internal/authz/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

type ExternalServicesStore interface {
	ListGitLabConnections(context.Context) ([]*types.GitLabConnection, error)
	ListGitHubConnections(context.Context) ([]*types.GitHubConnection, error)
	ListBitbucketServerConnections(context.Context) ([]*types.BitbucketServerConnection, error)
}

// ProvidersFromConfig returns the set of permission-related providers derived from the site config.
// It also returns any validation problems with the config, separating these into "serious problems"
// and "warnings". "Serious problems" are those that should make Sourcegraph set authz.allowAccessByDefault
// to false. "Warnings" are all other validation problems.
func ProvidersFromConfig(
	ctx context.Context,
	cfg *conf.Unified,
	s ExternalServicesStore,
) (
	allowAccessByDefault bool,
	providers []authz.Provider,
	seriousProblems []string,
	warnings []string,
) {
	allowAccessByDefault = true
	defer func() {
		if len(seriousProblems) > 0 {
			log15.Error("Repository authz config was invalid (errors are visible in the UI as an admin user, you should fix ASAP). Restricting access to repositories by default for now to be safe.", "seriousProblems", seriousProblems)
			allowAccessByDefault = false
		}
	}()

	if ghConns, err := s.ListGitHubConnections(ctx); err != nil {
		seriousProblems = append(seriousProblems, fmt.Sprintf("Could not load GitHub external service configs: %s", err))
	} else {
		ghProviders, ghProblems, ghWarnings := github.NewAuthzProviders(ghConns)
		providers = append(providers, ghProviders...)
		seriousProblems = append(seriousProblems, ghProblems...)
		warnings = append(warnings, ghWarnings...)
	}

	if glConns, err := s.ListGitLabConnections(ctx); err != nil {
		seriousProblems = append(seriousProblems, fmt.Sprintf("Could not load GitLab external service configs: %s", err))
	} else {
		glProviders, glProblems, glWarnings := gitlab.NewAuthzProviders(cfg, glConns)
		providers = append(providers, glProviders...)
		seriousProblems = append(seriousProblems, glProblems...)
		warnings = append(warnings, glWarnings...)
	}

	if bbsConns, err := s.ListBitbucketServerConnections(ctx); err != nil {
		seriousProblems = append(seriousProblems, fmt.Sprintf("Could not load Bitbucket Server external service configs: %s", err))
	} else {
		bbsProviders, bbsProblems, bbsWarnings := bitbucketserver.NewAuthzProviders(bbsConns)
		providers = append(providers, bbsProviders...)
		seriousProblems = append(seriousProblems, bbsProblems...)
		warnings = append(warnings, bbsWarnings...)
	}

	// 🚨 SECURITY: Warn the admin when both code host authz provider and the permissions user mapping are configured.
	if cfg.SiteConfiguration.PermissionsUserMapping != nil &&
		cfg.SiteConfiguration.PermissionsUserMapping.Enabled && len(providers) > 0 {
		serviceTypes := make([]string, len(providers))
		for i := range providers {
			serviceTypes[i] = strconv.Quote(providers[i].ServiceType())
		}
		msg := fmt.Sprintf(
			"The permissions user mapping (site configuration `permissions.userMapping`) cannot be enabled when %s authorization providers are in use. Blocking access to all repositories until the conflict is resolved.",
			strings.Join(serviceTypes, ", "))
		seriousProblems = append(seriousProblems, msg)
	}

	return allowAccessByDefault, providers, seriousProblems, warnings
}
