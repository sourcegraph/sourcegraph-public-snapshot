package repos

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/perforce"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A PerforceSource yields depots from a single Perforce connection configured
// in Sourcegraph via the external services configuration.
type PerforceSource struct {
	svc    *types.ExternalService
	config *schema.PerforceConnection
}

// NewPerforceSource returns a new PerforceSource from the given external
// service.
func NewPerforceSource(ctx context.Context, svc *types.ExternalService) (*PerforceSource, error) {
	rawConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	var c schema.PerforceConnection
	if err := jsonc.Unmarshal(rawConfig, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return newPerforceSource(svc, &c)
}

func newPerforceSource(svc *types.ExternalService, c *schema.PerforceConnection) (*PerforceSource, error) {
	return &PerforceSource{
		svc:    svc,
		config: c,
	}, nil
}

// CheckConnection tests the code host connection to make sure it works.
// For Perforce, it uses the host (p4.port), username (p4.user) and password (p4.passwd)
// from the code host configuration.
func (s PerforceSource) CheckConnection(ctx context.Context) error {
	// currently the only tool to use for connecting to the Perforce server
	// is the syncer because connecting requires the `p4` CLI binary, which is on `gitserver`
	syncer := server.PerforceDepotSyncer{}
	// ensure that the connection check won't go longer than 10 seconds, which should be plenty of time
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := syncer.CanConnect(timeoutCtx, s.config.P4Port, s.config.P4User, s.config.P4Passwd); err != nil {
		return errors.Wrap(err, "Unable to connect to the Perforce server")
	}
	return nil
}

// ListRepos returns all Perforce depots accessible to all connections
// configured in Sourcegraph via the external services configuration.
func (s PerforceSource) ListRepos(ctx context.Context, results chan SourceResult) {
	for _, depot := range s.config.Depots {
		// Tiny optimization: exit early if context has been canceled.
		if err := ctx.Err(); err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}

		u := url.URL{
			Scheme: "perforce",
			Host:   s.config.P4Port,
			Path:   depot,
			User:   url.UserPassword(s.config.P4User, s.config.P4Passwd),
		}
		p4Url, err := vcs.ParseURL(u.String())
		if err != nil {
			results <- SourceResult{Source: s, Err: err}
			continue
		}
		syncer := server.PerforceDepotSyncer{}
		// We don't need to provide repo name and use "" instead because p4 commands are
		// not recorded in the following `syncer.IsCloneable` call.
		if err := syncer.IsCloneable(ctx, "", p4Url); err == nil {
			results <- SourceResult{Source: s, Repo: s.makeRepo(depot)}
		} else {
			results <- SourceResult{Source: s, Err: err}
		}
	}
}

// composePerforceCloneURL composes a clone URL for a Perforce depot based on
// given information. e.g.
// perforce://ssl:111.222.333.444:1666//Sourcegraph/
func composePerforceCloneURL(host, depot string) string {
	cloneURL := url.URL{
		Scheme: "perforce",
		Host:   host,
		Path:   depot,
	}
	return cloneURL.String()
}

func (s PerforceSource) makeRepo(depot string) *types.Repo {
	if !strings.HasSuffix(depot, "/") {
		depot += "/"
	}
	name := strings.Trim(depot, "/")
	urn := s.svc.URN()

	cloneURL := composePerforceCloneURL(s.config.P4Port, depot)

	return &types.Repo{
		Name: reposource.PerforceRepoName(
			s.config.RepositoryPathPattern,
			name,
		),
		URI: string(reposource.PerforceRepoName(
			"",
			name,
		)),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          depot,
			ServiceType: extsvc.TypePerforce,
			ServiceID:   s.config.P4Port,
		},
		Private: true,
		Sources: map[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: cloneURL,
			},
		},
		Metadata: &perforce.Depot{
			Depot: depot,
		},
	}
}

// ExternalServices returns a singleton slice containing the external service.
func (s PerforceSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}
