package repos

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/goware/urlx"
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
//  1. Explicitly adding disabled repos that would have been added to an explicit exclude list.
//  2. Removing the deprecated initialRepositoryEnablement field.
//
// This migration must be rolled-out together with the UI changes that remove the admin's
// ability to explicitly enabled / disable individual repos.
func EnabledStateDeprecationMigration(sourcer Sourcer, clock func() time.Time, kinds ...string) Migration {
	return migrate(func(ctx context.Context, s Store) error {
		const prefix = "migrate.repos-enabled-state-deprecation:"

		es, err := s.ListExternalServices(ctx, StoreListExternalServicesArgs{
			Kinds: kinds,
		})

		if err != nil {
			return errors.Wrapf(err, "%s list-external-services", prefix)
		}

		srcs, err := sourcer(es...)
		if err != nil {
			return errors.Wrapf(err, "%s list-sources", prefix)
		}

		stored, err := s.ListRepos(ctx, StoreListReposArgs{
			Kinds: kinds,
		})

		if err != nil {
			return errors.Wrapf(err, "%s store.list-repos", prefix)
		}

		var disabled Repos
		for _, r := range stored {
			if !r.Enabled {
				disabled = append(disabled, r)
			}
		}

		all := srcs.ExternalServices()
		svcs := make(ExternalServices, 0, len(all))
		exclude := make(map[int64]*Repos, len(all))

		for _, e := range all {
			if e.ID != 0 {
				var excluded Repos
				exclude[e.ID] = &excluded
				svcs = append(svcs, e)
			}
		}

		if len(disabled) > 0 { // Only source and diff if there are disabled repos stored.
			var sourced Repos
			{
				ctx, cancel := context.WithTimeout(ctx, sourceTimeout)
				sourced, err = srcs.ListRepos(ctx)
				cancel()
			}

			if err != nil {
				return errors.Wrapf(err, "%s sources.list-repos", prefix)
			}

			diff := NewDiff(sourced, stored)
			for _, rs := range []Repos{diff.Added, diff.Modified, diff.Unmodified} {
				for _, r := range rs {
					if r.Enabled {
						continue
					}

					es := make(map[int64]*Repos, len(r.Sources))
					for _, si := range r.Sources {
						id := si.ExternalServiceID()
						if e := exclude[id]; e != nil {
							es[id] = e
						}
					}

					if len(es) == 0 {
						es = exclude
					}

					for _, e := range es {
						*e = append(*e, r)
					}
				}
			}
		}

		now := clock()
		for _, e := range svcs {
			if err = removeInitalRepositoryEnablement(e, now); err != nil {
				return errors.Wrapf(err, "%s remove-initial-repository-enablement", prefix)
			}

			if rs := exclude[e.ID]; rs != nil && len(*rs) > 0 {
				if err = e.Exclude(*rs...); err != nil {
					return errors.Wrapf(err, "%s exclude", prefix)
				}
				e.UpdatedAt = now

				log15.Info(prefix+" exclude", "service", e.DisplayName, "repos", len(*rs))
			}
		}

		if err = s.UpsertExternalServices(ctx, svcs...); err != nil {
			return errors.Wrapf(err, "%s upsert-external-services", prefix)
		}

		for _, r := range disabled {
			r.DeletedAt = now
		}

		if err = s.UpsertRepos(ctx, disabled...); err != nil {
			return errors.Wrapf(err, "%s upsert-repos", prefix)
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
		const prefix = "migrate.github-set-default-repository-query:"

		svcs, err := s.ListExternalServices(ctx, StoreListExternalServicesArgs{
			Kinds: []string{"github"},
		})

		if err != nil {
			return errors.Wrapf(err, "%s list-external-services", prefix)
		}

		now := clock()
		for _, svc := range svcs {
			var c schema.GitHubConnection
			if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
				return fmt.Errorf("%s external service id=%d config unmarshaling error: %s", prefix, svc.ID, err)
			}

			if len(c.RepositoryQuery) != 0 {
				continue
			}

			baseURL, err := url.Parse(c.Url)
			if err != nil {
				return errors.Wrapf(err, "%s parse-url", prefix)
			}

			_, githubDotCom := github.APIRoot(NormalizeBaseURL(baseURL))

			c.RepositoryQuery = append(c.RepositoryQuery, "affiliated")
			if !githubDotCom {
				c.RepositoryQuery = append(c.RepositoryQuery, "public")
			}

			edited, err := jsonc.Edit(svc.Config, c.RepositoryQuery, "repositoryQuery")
			if err != nil {
				return errors.Wrapf(err, "%s edit-json", prefix)
			}

			svc.Config = edited
			svc.UpdatedAt = now
		}

		if err = s.UpsertExternalServices(ctx, svcs...); err != nil {
			return errors.Wrapf(err, "%s upsert-external-services", prefix)
		}

		return nil
	})
}

// GitLabSetDefaultProjectQueryMigration returns a Migration that changes all
// configurations of GitLab external services which have an empty "projectQuery"
// migration to its explicit default.
func GitLabSetDefaultProjectQueryMigration(clock func() time.Time) Migration {
	return migrate(func(ctx context.Context, s Store) error {
		const prefix = "migrate.gitlab-set-default-project-query:"

		svcs, err := s.ListExternalServices(ctx, StoreListExternalServicesArgs{
			Kinds: []string{"gitlab"},
		})

		if err != nil {
			return errors.Wrapf(err, "%s list-external-services", prefix)
		}

		now := clock()
		for _, svc := range svcs {
			var c schema.GitLabConnection
			if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
				return fmt.Errorf("%s  external service id=%d config unmarshaling error: %s", prefix, svc.ID, err)
			}

			if len(c.ProjectQuery) != 0 {
				continue
			}

			c.ProjectQuery = append(c.ProjectQuery, "?membership=true")

			edited, err := jsonc.Edit(svc.Config, c.ProjectQuery, "projectQuery")
			if err != nil {
				return errors.Wrapf(err, "%s edit-json", prefix)
			}

			svc.Config = edited
			svc.UpdatedAt = now
		}

		if err = s.UpsertExternalServices(ctx, svcs...); err != nil {
			return errors.Wrapf(err, "%s upsert-external-services", prefix)
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
		const prefix = "migrate.bitbucketserver-set-default-repository-query:"

		svcs, err := s.ListExternalServices(ctx, StoreListExternalServicesArgs{
			Kinds: []string{"bitbucketserver"},
		})

		if err != nil {
			return errors.Wrapf(err, "%s list-external-services", prefix)
		}

		now := clock()
		for _, svc := range svcs {
			var c schema.BitbucketServerConnection
			if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
				return fmt.Errorf("%s  external service id=%d config unmarshaling error: %s", prefix, svc.ID, err)
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
				return errors.Wrapf(err, "%s edit-json", prefix)
			}

			svc.Config = edited
			svc.UpdatedAt = now
		}

		if err = s.UpsertExternalServices(ctx, svcs...); err != nil {
			return errors.Wrapf(err, "%s upsert-external-services", prefix)
		}

		return nil
	})
}

// BitbucketServerUsernameMigration returns a Migration that changes all
// configurations of BitbucketServer external services to explicitly have the
// `username` setting set to the user defined in the `url`, if any.
// This will only happen if the `username` fields is empty or unset.
func BitbucketServerUsernameMigration(clock func() time.Time) Migration {
	return migrate(func(ctx context.Context, s Store) error {
		const prefix = "migrate.bitbucketserver-username-migration:"

		svcs, err := s.ListExternalServices(ctx, StoreListExternalServicesArgs{
			Kinds: []string{"bitbucketserver"},
		})

		if err != nil {
			return errors.Wrapf(err, "%s list-external-services", prefix)
		}

		now := clock()
		for _, svc := range svcs {
			var c schema.BitbucketServerConnection
			if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
				return errors.Errorf("%s  external service id=%d config unmarshaling error: %s", prefix, svc.ID, err)
			}

			if c.Username != "" {
				continue
			}

			u, err := urlx.Parse(c.Url)
			if err != nil {
				return errors.Wrapf(err, "%s parse-url", prefix)
			}

			username := u.User.Username()
			if username == "" {
				continue
			}

			edited, err := jsonc.Edit(svc.Config, username, "username")
			if err != nil {
				return errors.Wrapf(err, "%s edit-json", prefix)
			}

			svc.Config = edited
			svc.UpdatedAt = now
		}

		if err = s.UpsertExternalServices(ctx, svcs...); err != nil {
			return errors.Wrapf(err, "%s upsert-external-services", prefix)
		}

		return nil
	})
}

// AWSCodeCommitSetBogusGitCredentialsMigration returns a Migration that
// changes all configurations of AWS CodeCommit external services to have the
// `gitCredentials` setting set to bogus defaults. This will only happen if the
// `gitCredentials` fields is empty or missing either `username` or `password`.
//
// This is done so that repo-updater can boot up without an "invalid config"
// error and accept new external service configuration syncs, which allows the
// user to fix the wrong credentials.
func AWSCodeCommitSetBogusGitCredentialsMigration(clock func() time.Time) Migration {

	return migrate(func(ctx context.Context, s Store) error {
		const (
			prefix = "migrate.aws-codecommit-set-bogus-git-credentials:"

			defaultUsername = "insert-git-credentials-username-here"
			defaultPassword = "insert-git-credentials-password-here"
		)

		svcs, err := s.ListExternalServices(ctx, StoreListExternalServicesArgs{
			Kinds: []string{"awscodecommit"},
		})

		if err != nil {
			return errors.Wrapf(err, "%s list-external-services", prefix)
		}

		now := clock()
		for _, svc := range svcs {
			var c schema.AWSCodeCommitConnection
			if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
				return errors.Errorf("%s external service id=%d config unmarshaling error: %s", prefix, svc.ID, err)
			}

			gitCredentials := c.GitCredentials

			if gitCredentials.Username != "" && gitCredentials.Password != "" {
				continue
			}

			if gitCredentials.Username == "" {
				gitCredentials.Username = defaultUsername
			}

			if gitCredentials.Password == "" {
				gitCredentials.Password = defaultPassword
			}

			edited, err := jsonc.Edit(svc.Config, gitCredentials, "gitCredentials")
			if err != nil {
				return errors.Wrapf(err, "%s edit-json", prefix)
			}

			svc.Config = edited
			svc.UpdatedAt = now
		}

		if err = s.UpsertExternalServices(ctx, svcs...); err != nil {
			return errors.Wrapf(err, "%s upsert-external-services", prefix)
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
