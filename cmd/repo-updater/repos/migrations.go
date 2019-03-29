package repos

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/lib/pq"
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
func GithubReposEnabledStateDeprecationMigration(sourcer Sourcer, clock func() time.Time) Migration {
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
						} else {
							return fmt.Errorf("external service with id=%d does not exist", id)
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
				if err = e.svc.ExcludeGithubRepos(e.exclude...); err != nil {
					return errors.Wrapf(err, "%s.exclude", prefix)
				}
				e.svc.UpdatedAt = now
			}

			if len(e.include) > 0 {
				if err = e.svc.IncludeGithubRepos(e.include...); err != nil {
					return errors.Wrapf(err, "%s.include", prefix)
				}
				e.svc.UpdatedAt = now
			}

		}

		return s.UpsertExternalServices(ctx, upserts...)
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
	return transactional(func(ctx context.Context, s Store) error {
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
	return transactional(func(ctx context.Context, s Store) error {
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
			if err = transact(ctx, tr, m); err == nil || !isRetryable(err) {
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

func isRetryable(err error) bool {
	switch e := errors.Cause(err).(type) {
	case *pq.Error:
		switch e.Code.Class() {
		case "40":
			// Class 40 — Transaction Rollback
			// 40000	transaction_rollback
			// 40002	transaction_integrity_constraint_violation
			// 40001	serialization_failure
			// 40003	statement_completion_unknown
			// 40P01	deadlock_detected
			return true
		}
	}
	return false
}
