package repos

import (
	"context"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/perforce"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A PerforceSource yields depots from a single Perforce connection configured
// in Sourcegraph via the external services configuration.
type PerforceSource struct {
	gitserverClient gitserver.Client
	svc             *types.ExternalService
	config          *schema.PerforceConnection
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
	return newPerforceSource(gitserver.NewClient("repos.perforcesource"), svc, &c)
}

func newPerforceSource(gitserverClient gitserver.Client, svc *types.ExternalService, c *schema.PerforceConnection) (*PerforceSource, error) {
	return &PerforceSource{
		svc:             svc,
		config:          c,
		gitserverClient: gitserverClient,
	}, nil
}

// CheckConnection tests the code host connection to make sure it works.
// For Perforce, it uses the host (p4.port), username (p4.user) and password (p4.passwd)
// from the code host configuration.
func (s PerforceSource) CheckConnection(ctx context.Context) error {
	gclient := gitserver.NewClient("perforce.connectioncheck")
	conn := protocol.PerforceConnectionDetails{
		P4Port:   s.config.P4Port,
		P4User:   s.config.P4User,
		P4Passwd: s.config.P4Passwd,
	}
	err := gclient.CheckPerforceCredentials(ctx, conn)
	if err != nil {
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

		conn := protocol.PerforceConnectionDetails{
			P4Port:   s.config.P4Port,
			P4User:   s.config.P4User,
			P4Passwd: s.config.P4Passwd,
		}
		err := s.gitserverClient.IsPerforcePathCloneable(ctx, conn, depot)
		if err != nil {
			results <- SourceResult{Source: s, Err: errors.Wrap(err, "checking if perforce path is cloneable")}
			continue
		}

		results <- SourceResult{Source: s, Repo: s.makeRepo(depot)}
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
