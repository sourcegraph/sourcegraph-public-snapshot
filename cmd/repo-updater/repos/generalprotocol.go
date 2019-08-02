package repos

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf/reposource"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/generalprotocol"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
	"gopkg.in/inconshreveable/log15.v2"
)

// A GeneralProtocolSource yields repositories from a single general protocol connection configured
// in Sourcegraph via the external services configuration.
type GeneralProtocolSource struct {
	svc    *ExternalService
	config *schema.GeneralProtocolConnection
	client *generalprotocol.Client
	info   *generalprotocol.Info
}

// NewGeneralProtocolSource returns a new GeneralProtocolSource from the given external service.
func NewGeneralProtocolSource(svc *ExternalService, cf *httpcli.Factory) (*GeneralProtocolSource, error) {
	var c schema.GeneralProtocolConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, fmt.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return newGeneralProtocolSource(svc, &c, cf)
}

func newGeneralProtocolSource(svc *ExternalService, c *schema.GeneralProtocolConnection, cf *httpcli.Factory) (*GeneralProtocolSource, error) {
	endpointURL, err := url.Parse(c.Endpoint)
	if err != nil {
		return nil, err
	}
	endpointURL = extsvc.NormalizeBaseURL(endpointURL)

	if cf == nil {
		cf = NewHTTPClientFactory()
	}

	cli, err := cf.Doer()
	if err != nil {
		return nil, err
	}

	client := generalprotocol.NewClient(endpointURL, cli)
	client.Token = c.Token
	client.Username = c.Username
	client.Password = c.Password

	return &GeneralProtocolSource{
		svc:    svc,
		config: c,
		client: client,
	}, nil
}

// ListRepos returns all General Protocol repositories accessible to all connections configured
// in Sourcegraph via the external services configuration.
func (s GeneralProtocolSource) ListRepos(ctx context.Context) (repos []*Repo, err error) {
	rs, err := s.listAllRepos(ctx)
	for _, r := range rs {
		repos = append(repos, s.makeRepo(r))
	}
	return repos, err
}

// ExternalServices returns a singleton slice containing the external service.
func (s GeneralProtocolSource) ExternalServices() ExternalServices {
	return ExternalServices{s.svc}
}

func (s GeneralProtocolSource) makeRepo(r *generalprotocol.Repo) *Repo {
	host, err := url.Parse(s.config.Url)
	if err != nil {
		// This should never happen
		panic(errors.Errorf("malformed General Protocol config, invalid URL: %q, error: %s", s.config.Url, err))
	}
	host = extsvc.NormalizeBaseURL(host)

	urn := s.svc.URN()
	return &Repo{
		Name: string(reposource.GeneralProtocolRepoName(
			s.config.RepositoryPathPattern,
			host.Hostname(),
			r.FullName,
		)),
		URI: string(reposource.GeneralProtocolRepoName(
			"",
			host.Hostname(),
			r.FullName,
		)),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          r.ID,
			ServiceType: generalprotocol.ServiceType,
			ServiceID:   host.String(),
		},
		Description: r.Description,
		Fork:        r.Parent != nil,
		Enabled:     true,
		Sources: map[string]*SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: s.authenticatedRemoteURL(r),
			},
		},
		Metadata: r,
	}
}

// authenticatedRemoteURL returns the repository's Git remote URL with the configured
// credentials inserted in the URL userinfo.
func (s *GeneralProtocolSource) authenticatedRemoteURL(repo *generalprotocol.Repo) string {
	fallbackURL := (&url.URL{
		Scheme: "https",
		Host:   s.config.Url,
		Path:   "/" + repo.FullName,
	}).String()

	if s.config.GitURLType == "ssh" {
		sshURL, err := repo.SSHLink()
		if err != nil {
			log15.Warn("Error obtaining General Protocol repository Git remote URL.", "url", repo.Links, "error", err)
			return fallbackURL
		}

		return sshURL
	}

	httpURL, err := repo.HTTPLink()
	if err != nil {
		log15.Warn("Error obtaining General Protocol repository Git remote URL.", "url", repo.Links, "error", err)
		return fallbackURL
	}
	u, err := url.Parse(httpURL)
	if err != nil {
		log15.Warn("Error parsing General Protocol repository Git remote URL.", "url", httpURL, "error", err)
		return fallbackURL
	}

	if s.config.Token != "" {
		u.User = url.User(s.config.Token)
		return u.String()
	}

	password := s.config.Token
	if password == "" {
		password = s.config.Password
	}
	u.User = url.UserPassword(s.config.Username, password)
	return u.String()
}

func (s *GeneralProtocolSource) listAllRepos(ctx context.Context) ([]*generalprotocol.Repo, error) {
	var err error
	s.info, err = s.client.Info(ctx)
	if err != nil {
		return nil, errors.Errorf("request info: %v", err)
	}

	type batch struct {
		repos []*generalprotocol.Repo
		err   error
	}

	ch := make(chan batch)

	var wg sync.WaitGroup

	// List all repositories belonging to the account
	wg.Add(1)
	go func() {
		defer wg.Done()

		page := &generalprotocol.PageToken{PageLen: s.info.MaxPageLen}
		var err error
		var repos []*generalprotocol.Repo
		for !page.IsLastPage {
			if repos, page, err = s.client.Repos(ctx, page); err != nil {
				ch <- batch{err: errors.Wrapf(err, "generalprotocol.repos: page=%+v", page)}
				break
			}

			ch <- batch{repos: repos}
		}
	}()

	// List all repositories of organizations selected that the account has access to
	wg.Add(1)
	go func() {
		defer wg.Done()

		for _, o := range s.config.Orgs {
			page := &generalprotocol.PageToken{PageLen: s.info.MaxPageLen}
			var err error
			var repos []*generalprotocol.Repo
			for !page.IsLastPage {
				if repos, page, err = s.client.UserRepos(ctx, page, o); err != nil {
					ch <- batch{err: errors.Wrapf(err, "generalprotocol.orgs: item=%q, page=%+v", o, page)}
					break
				}

				ch <- batch{repos: repos}
			}
		}
	}()

	go func() {
		wg.Wait()
		close(ch)
	}()

	seen := make(map[string]bool)
	errs := new(multierror.Error)
	var repos []*generalprotocol.Repo

	for r := range ch {
		if r.err != nil {
			errs = multierror.Append(errs, r.err)
		}

		for _, repo := range r.repos {
			// Discard non-Git repositories
			if repo.SCM != "git" {
				continue
			}

			if !seen[repo.ID] {
				repos = append(repos, repo)
				seen[repo.ID] = true
			}
		}
	}

	return repos, errs.ErrorOrNil()
}
