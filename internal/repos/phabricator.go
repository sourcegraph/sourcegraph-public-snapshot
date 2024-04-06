package repos

import (
	"context"
	"sync"

	"github.com/goware/urlx"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/phabricator"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A PhabricatorSource yields repositories from a single Phabricator connection configured
// in Sourcegraph via the external services configuration.
type PhabricatorSource struct {
	svc  *types.ExternalService
	conn *schema.PhabricatorConnection
	cf   *httpcli.Factory

	mu     sync.Mutex
	cli    *phabricator.Client
	logger log.Logger
}

// NewPhabricatorSource returns a new PhabricatorSource from the given external service.
func NewPhabricatorSource(ctx context.Context, logger log.Logger, svc *types.ExternalService, cf *httpcli.Factory) (*PhabricatorSource, error) {
	rawConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	var c schema.PhabricatorConnection
	if err := jsonc.Unmarshal(rawConfig, &c); err != nil {
		return nil, errors.Wrapf(err, "external service id=%d config error", svc.ID)
	}
	return &PhabricatorSource{logger: logger, svc: svc, conn: &c, cf: cf}, nil
}

// CheckConnection at this point assumes availability and relies on errors returned
// from the subsequent calls. This is going to be expanded as part of issue #44683
// to actually only return true if the source can serve requests.
func (s *PhabricatorSource) CheckConnection(ctx context.Context) error {
	return nil
}

// ListRepos returns all Phabricator repositories accessible to all connections configured
// in Sourcegraph via the external services configuration.
func (s *PhabricatorSource) ListRepos(ctx context.Context, results chan SourceResult) {
	cli, err := s.client(ctx)
	if err != nil {
		results <- SourceResult{Source: s, Err: err}
		return
	}

	cursor := &phabricator.Cursor{Limit: 100, Order: "oldest"}
	for {
		var page []*phabricator.Repo
		page, cursor, err = cli.ListRepos(ctx, phabricator.ListReposArgs{Cursor: cursor})
		if err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}

		for _, r := range page {
			if r.VCS != "git" || r.Status == "inactive" {
				continue
			}

			repo, err := s.makeRepo(r)
			if err != nil {
				results <- SourceResult{Source: s, Err: err}
				return
			}
			results <- SourceResult{Source: s, Repo: repo}
		}

		if cursor.After == "" {
			break
		}
	}
}

// ExternalServices returns a singleton slice containing the external service.
func (s *PhabricatorSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func (s *PhabricatorSource) makeRepo(repo *phabricator.Repo) (*types.Repo, error) {
	var external []*phabricator.URI
	builtin := make(map[string]*phabricator.URI)

	for _, u := range repo.URIs {
		if u.Disabled || u.Normalized == "" {
			continue
		} else if u.BuiltinIdentifier != "" {
			builtin[u.BuiltinProtocol+"+"+u.BuiltinIdentifier] = u
		} else {
			external = append(external, u)
		}
	}

	var name string
	if len(external) > 0 {
		name = external[0].Normalized
	}

	var cloneURL string
	for _, alt := range [...]struct {
		protocol, identifier string
	}{ // Ordered by priority.
		{"https", "shortname"},
		{"https", "callsign"},
		{"https", "id"},
		{"ssh", "shortname"},
		{"ssh", "callsign"},
		{"ssh", "id"},
	} {
		if u, ok := builtin[alt.protocol+"+"+alt.identifier]; ok {
			cloneURL = u.Effective

			if name == "" {
				name = u.Normalized
			}
		}
	}

	if cloneURL == "" {
		s.logger.Warn("unable to construct clone URL for repo", log.String("name", name), log.String("phabricator_id", repo.PHID))
	}

	if name == "" {
		return nil, errors.Errorf("no canonical name available for repo with id=%v", repo.PHID)
	}

	serviceID, err := urlx.NormalizeString(s.conn.Url)
	if err != nil {
		// Should never happen. URL must be validated on input.
		panic(err)
	}

	urn := s.svc.URN()
	return &types.Repo{
		Name: api.RepoName(name),
		URI:  name,
		ExternalRepo: api.ExternalRepoSpec{
			ID:          repo.PHID,
			ServiceType: extsvc.TypePhabricator,
			ServiceID:   serviceID,
		},
		Sources: map[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: cloneURL,
				// TODO(tsenart): We need a way for admins to specify which URI to
				// use as a CloneURL. Do they want to use https + shortname, git + callsign
				// an external URI that's mirrored or observed, etc.
				// This must be figured out when starting to integrate the new Syncer with this
				// source.
			},
		},
		Metadata: repo,
		Private:  !s.svc.Unrestricted,
	}, nil
}

// client initialises the phabricator.Client if it isn't initialised yet.
// This is done lazily instead of in NewPhabricatorSource so that we have
// access to the context.Context passed in via ListRepos.
func (s *PhabricatorSource) client(ctx context.Context) (*phabricator.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cli != nil {
		return s.cli, nil
	}

	hc, err := s.cf.Doer()
	if err != nil {
		return nil, err
	}

	s.cli, err = phabricator.NewClient(ctx, s.conn.Url, s.conn.Token, hc)
	return s.cli, err
}
