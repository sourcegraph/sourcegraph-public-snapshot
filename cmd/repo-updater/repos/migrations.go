package repos

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// A Migration performs a data migration in the given Store,
// returning an error in case of failure.
type Migration func(context.Context, Store) error

// Run is an utility method to aid readability of calling code.
func (m Migration) Run(ctx context.Context, s Store) error {
	return m(ctx, s)
}

// EnabledStateDeprecationMigration returns a Migration that changes
// existing external services to maintain the same set of mirrored repos
// without recourse to the now deprecated enabled column of a repository.
//
// This is done by:
//  1. Explicitly adding enabled repos that would have been deleted to an explicit include list.
//  2. Explicitly adding disabled repos that would have been added to an explicit exclude list.
//  3. Removing the deprecated initialRepositoryEnablement field.
//
// This migration must be rolled-out together with the UI changes that remove the admin's
// ability to explicitly enabled / disable individual repos.
func EnabledStateDeprecationMigration(sourcer Sourcer, clock func() time.Time, kinds ...string) Migration {
	return migrate(func(ctx context.Context, s Store) error {
		const prefix = "migrate.repos-enabled-state-deprecation"

		es, err := s.ListExternalServices(ctx, kinds...)
		if err != nil {
			return errors.Wrapf(err, "%s.list-external-services", prefix)
		}

		srcs, err := sourcer(es...)
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

		stored, err := s.ListRepos(ctx, StoreListReposArgs{
			Kinds: kinds,
		})

		if err != nil {
			return errors.Wrapf(err, "%s.store.list-repos", prefix)
		}

		type service struct {
			svc     *ExternalService
			include Repos
			exclude Repos
		}

		all := srcs.ExternalServices()
		svcs := make(map[int64]*service, len(all))
		upserts := make(ExternalServices, 0, len(all))

		for _, e := range all {
			// Skip any injected sources that are not persisted.
			if e.ID != 0 {
				svcs[e.ID] = &service{svc: e}
				upserts = append(upserts, e)
			}
		}

		group := func(pred func(*Repo) bool, bucket func(*service) *Repos, repos ...Repos) error {
			for _, rs := range repos {
				for _, r := range rs {
					if !pred(r) {
						continue
					}

					es := make(map[int64]*service, len(r.Sources))
					for _, si := range r.Sources {
						id := si.ExternalServiceID()
						if e := svcs[id]; e != nil {
							es[id] = e
						}
					}

					if len(es) == 0 {
						es = svcs
					}

					for _, e := range es {
						b := bucket(e)
						*b = append(*b, r)
					}
				}
			}

			return nil
		}

		diff := NewDiff(sourced, stored)

		err = group(
			func(r *Repo) bool { return !r.Enabled },
			func(s *service) *Repos { return &s.exclude },
			diff.Added, diff.Modified, diff.Unmodified,
		)

		if err != nil {
			return err
		}

		err = group(
			func(r *Repo) bool { return r.Enabled },
			func(s *service) *Repos { return &s.include },
			diff.Deleted,
		)

		if err != nil {
			return err
		}

		now := clock()
		for _, e := range svcs {
			if err = removeInitalRepositoryEnablement(e.svc, now); err != nil {
				return errors.Wrapf(err, "%s.remove-initial-repository-enablement", prefix)
			}

			if len(e.exclude) > 0 {
				if err = e.svc.Exclude(e.exclude...); err != nil {
					return errors.Wrapf(err, "%s.exclude", prefix)
				}
				e.svc.UpdatedAt = now

				log15.Info(prefix+".exclude", "service", e.svc.DisplayName, "repos", len(e.exclude))
			}

			if len(e.include) > 0 {
				if err = e.svc.Include(e.include...); err != nil {
					return errors.Wrapf(err, "%s.include", prefix)
				}
				e.svc.UpdatedAt = now

				log15.Info(prefix+".include", "service", e.svc.DisplayName, "repos", len(e.include))
			}

		}

		if err = s.UpsertExternalServices(ctx, upserts...); err != nil {
			return errors.Wrapf(err, "%s.upsert-external-services", prefix)
		}

		var deleted Repos
		for _, r := range stored {
			if !r.Enabled {
				r.DeletedAt = now
				r.Enabled = true
				deleted = append(deleted, r)
			}
		}

		if err = s.UpsertRepos(ctx, deleted...); err != nil {
			return errors.Wrapf(err, "%s.upsert-repos", prefix)
		}

		return nil
	})
}

func removeInitalRepositoryEnablement(svc *ExternalService, ts time.Time) error {
	edited, err := jsonc.Remove(svc.Config, "initialRepositoryEnablement")
	if err != nil {
		return err
	}

	if edited != svc.Config {
		svc.Config = edited
		svc.UpdatedAt = ts
	}

	return nil
}

// GithubSetDefaultRepositoryQueryMigration returns a Migration that changes all
// configurations of GitHub external services which have an empty "repositoryQuery"
// migration to its explicit default.
func GithubSetDefaultRepositoryQueryMigration(clock func() time.Time) Migration {
	return migrate(func(ctx context.Context, s Store) error {
		const prefix = "migrate.github-set-default-repository-query"

		svcs, err := s.ListExternalServices(ctx, "github")
		if err != nil {
			return errors.Wrapf(err, "%s.list-external-services", prefix)
		}

		now := clock()
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
			svc.UpdatedAt = now
		}

		if err = s.UpsertExternalServices(ctx, svcs...); err != nil {
			return errors.Wrapf(err, "%s.upsert-external-services", prefix)
		}

		return nil
	})
}

// GitLabSetDefaultProjectQueryMigration returns a Migration that changes all
// configurations of GitLab external services which have an empty "projectQuery"
// migration to its explicit default.
func GitLabSetDefaultProjectQueryMigration(clock func() time.Time) Migration {
	return migrate(func(ctx context.Context, s Store) error {
		const prefix = "migrate.gitlab-set-default-project-query"

		svcs, err := s.ListExternalServices(ctx, "gitlab")
		if err != nil {
			return errors.Wrapf(err, "%s.list-external-services", prefix)
		}

		now := clock()
		for _, svc := range svcs {
			var c schema.GitLabConnection
			if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
				return fmt.Errorf("%s: external service id=%d config unmarshaling error: %s", prefix, svc.ID, err)
			}

			if len(c.ProjectQuery) != 0 {
				continue
			}

			c.ProjectQuery = append(c.ProjectQuery, "?membership=true")

			edited, err := jsonc.Edit(svc.Config, c.ProjectQuery, "projectQuery")
			if err != nil {
				return errors.Wrapf(err, "%s.edit-json", prefix)
			}

			svc.Config = edited
			svc.UpdatedAt = now
		}

		if err = s.UpsertExternalServices(ctx, svcs...); err != nil {
			return errors.Wrapf(err, "%s.upsert-external-services", prefix)
		}

		return nil
	})
}

// BitbucketServerSetDefaultRepositoryQueryMigration returns a Migration that changes all
// configurations of BitbucketServer external services to explicitly have the new
// `repositoryQuery` setting set to a value that results in the semantically equivalent
// behaviour of mirroring all repos accessible to the configured token.
func BitbucketServerSetDefaultRepositoryQueryMigration(clock func() time.Time) Migration {
	return migrate(func(ctx context.Context, s Store) error {
		const prefix = "migrate.bitbucketserver-set-default-repository-query"

		svcs, err := s.ListExternalServices(ctx, "bitbucketserver")
		if err != nil {
			return errors.Wrapf(err, "%s.list-external-services", prefix)
		}

		now := clock()
		for _, svc := range svcs {
			var c schema.BitbucketServerConnection
			if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
				return fmt.Errorf("%s: external service id=%d config unmarshaling error: %s", prefix, svc.ID, err)
			}

			if len(c.RepositoryQuery) != 0 {
				continue
			}

			c.RepositoryQuery = append(c.RepositoryQuery,
				"?visibility=private",
				"?visibility=public",
			)

			edited, err := jsonc.Edit(svc.Config, c.RepositoryQuery, "repositoryQuery")
			if err != nil {
				return errors.Wrapf(err, "%s.edit-json", prefix)
			}

			svc.Config = edited
			svc.UpdatedAt = now
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

// migrate wraps a Migration with transactional and retries.
func migrate(m Migration) Migration {
	return func(ctx context.Context, s Store) (err error) {
		tr, ok := s.(Transactor)
		if !ok {
			return ErrNoTransactor
		}

		const wait = 5 * time.Second
		for {
			if err = transact(ctx, tr, m); err == nil {
				return nil
			}

			log15.Error("migrate", "error", err, "waiting", wait)
			time.Sleep(wait)
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
