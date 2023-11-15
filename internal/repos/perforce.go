package repos

import (
	"context"
	"net/url"
	"os"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/perforce"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
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
	// since CheckConnection is called from the frontend, we can't rely on the `p4` executable
	// being available, so we need to make an RPC call to `gitserver`, where it is available.
	// Use what is for us a "no-op" `p4` command that should always succeed.
	gclient := gitserver.NewClient()
	rc, _, err := gclient.P4Exec(ctx, s.config.P4Port, s.config.P4User, s.config.P4Passwd, "users")
	if err != nil {
		return errors.Wrap(err, "Unable to connect to the Perforce server")
	}
	rc.Close()
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
		tmpDir, err := os.MkdirTemp("", "perforce-syncer")
		if err != nil {
			log.Scoped("repos.source.perforce", "").Error("failed to create tmpDir", log.Error(err))
		}
		syncer := server.PerforceDepotSyncer{P4Home: tmpDir}
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
func composePerforceCloneURL(host, depot, username, password string) string {
	cloneURL := url.URL{
		Scheme: "perforce",
		Host:   host,
		Path:   depot,
	}
	if username != "" && password != "" {
		cloneURL.User = url.UserPassword(username, password)
	}
	return cloneURL.String()
}

func (s PerforceSource) makeRepo(depot string) *types.Repo {
	if !strings.HasSuffix(depot, "/") {
		depot += "/"
	}
	name := strings.Trim(depot, "/")
	urn := s.svc.URN()

	cloneURL := composePerforceCloneURL(s.config.P4Port, depot, "", "")

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
