package repos

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A Migration performs a data migration in the given Store,
// returning an error in case of failure.
type Migration func(context.Context, Store) error

// Run is an utility method to aid readability of calling code.
func (m Migration) Run(ctx context.Context, s Store) error {
	return m(ctx, s)
}

// GithubReposEnabledStateDeprecationMigration returns a Migration that changes
// existing Github external services to maintain the same set of mirrored repos
// without recourse to the now deprecated enabled column of a repository.
//
// This is done by:
//  1. Explicitly adding enabled repos that would have been deleted to github.repos
//  2. Explicitly adding disabled repos that would have been added to github.exclude
//  3. Removing the deprecated github.initialRepositoryEnablement field.
//
// This migration must be rolled-out together with the UI changes that remove the admin's
// ability to explicitly enabled / disable individual repos.
func GithubReposEnabledStateDeprecationMigration(sourcer Sourcer) Migration {
	return transactional(func(ctx context.Context, s Store) error {
		const prefix = "migrate.github-repos-enabled-state-deprecation"

		githubs, err := s.ListExternalServices(ctx, "github")
		if err != nil {
			return errors.Wrapf(err, "%s.list-external-services", prefix)
		}

		srcs, err := sourcer(githubs...)
		if err != nil {
			return errors.Wrapf(err, "%s.list-sources", prefix)
		}

		var sourced Repos
		{
			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			sourced, err = srcs.ListRepos(ctx)
			cancel()
		}

		if err != nil {
			return errors.Wrapf(err, "%s.sources.list-repos", prefix)
		}

		stored, err := s.ListRepos(ctx, "github")
		if err != nil {
			return errors.Wrapf(err, "%s.store.list-repos", prefix)
		}

		diff := NewDiff(sourced, stored)

		var disabled Repos
		for _, rs := range [...]Repos{diff.Added, diff.Modified, diff.Unmodified} {
			for _, r := range rs {
				if !r.Enabled {
					disabled = append(disabled, r)
				}
			}
		}

		var enabled Repos
		for _, r := range diff.Deleted {
			if r.Enabled {
				enabled = append(enabled, r)
			}
		}

		all := srcs.ExternalServices()
		svcs := make(map[int64]*ExternalService, len(all))
		for _, svc := range all {
			svcs[svc.ID] = svc
			if err = removeInitalRepositoryEnablement(svc); err != nil {
				return errors.Wrapf(err, "%s.remove-initial-repository-enablement", prefix)
			}
		}

		for _, r := range disabled {
			var es ExternalServices
			for _, si := range r.Sources {
				if svc := svcs[si.ExternalServiceID()]; svc != nil {
					es = append(es, svc)
				}
			}

			if len(es) == 0 {
				// If the repo was deleted, and it is still deleted, it has no
				// sources stored. So we add it to all Github external services
				// just to be sure.
				es = all
			}

			for _, svc := range es {
				if err := svc.ExcludeGithubRepos(r); err != nil {
					return errors.Wrapf(err, "%s.disabled", prefix)
				}
			}
		}

		for _, r := range enabled {
			var es ExternalServices
			for _, si := range r.Sources {
				if svc := svcs[si.ExternalServiceID()]; svc != nil {
					es = append(es, svc)
				}
			}

			for _, svc := range es {
				if err := svc.IncludeGithubRepos(r); err != nil {
					return errors.Wrapf(err, "%s.enabled", prefix)
				}
			}
		}

		return s.UpsertExternalServices(ctx, all...)
	})
}

func removeInitalRepositoryEnablement(svc *ExternalService) error {
	edited, err := jsonc.Remove(svc.Config, "initialRepositoryEnablement")
	if err != nil {
		return err
	}

	svc.Config = edited
	return nil
}

// GithubSetDefaultRepositoryQueryMigration returns a Migration that changes all
// configurations of GitHub external services which have an empty "repositoryQuery"
// migration to its explicit default.
func GithubSetDefaultRepositoryQueryMigration() Migration {
	return transactional(func(ctx context.Context, s Store) error {
		const prefix = "migrate.github-set-default-repository-query"

		svcs, err := s.ListExternalServices(ctx, "github")
		if err != nil {
			return errors.Wrapf(err, "%s.list-external-services", prefix)
		}

		for _, svc := range svcs {
			var c schema.GitHubConnection
			if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
				return fmt.Errorf("%s: external service id=%d config unmarshaling error: %s", prefix, svc.ID, err)
			}

			if len(c.RepositoryQuery) != 0 {
				continue
			}

			baseURL, err := url.Parse(c.Url)
			if err != nil {
				return errors.Wrapf(err, "%s.parse-url", prefix)
			}

			_, githubDotCom := github.APIRoot(NormalizeBaseURL(baseURL))

			c.RepositoryQuery = append(c.RepositoryQuery, "affiliated")
			if !githubDotCom {
				c.RepositoryQuery = append(c.RepositoryQuery, "public")
			}

			edited, err := jsonc.Edit(svc.Config, c.RepositoryQuery, "repositoryQuery")
			if err != nil {
				return errors.Wrapf(err, "%s.edit-json", prefix)
			}

			svc.Config = edited
		}

		if err = s.UpsertExternalServices(ctx, svcs...); err != nil {
			return errors.Wrapf(err, "%s.upsert-external-services", prefix)
		}

		return nil
	})
}

// ErrNoTransactor is returned by a Migration returned by
// NewTxMigration when it takes in a Store that can't be
// interface upgraded to a Transactor.
var ErrNoTransactor = errors.New("Store is not a Transactor")

// transactional wraps a Migration with transactional semantics. It retries
// the migration when a serialization transactional error is returned
// (i.e. ERROR: could not serialize access due to concurrent update)
func transactional(m Migration) Migration {
	return func(ctx context.Context, s Store) (err error) {
		tr, ok := s.(Transactor)
		if !ok {
			return ErrNoTransactor
		}

		for {
			if err = transact(ctx, tr, m); err == nil || !isSerializationError(err) {
				return err
			}
		}
	}
}

func transact(ctx context.Context, tr Transactor, m Migration) (err error) {
	var tx TxStore
	if tx, err = tr.Transact(ctx); err != nil {
		return err
	}

	defer tx.Done(&err)

	return m(ctx, tx)
}

func isSerializationError(err error) bool {
	return strings.Contains(err.Error(),
		"could not serialize access due to concurrent update")
}
