package repos

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/jsonx"
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

			edits, _, err := jsonx.ComputePropertyEdit(svc.Config,
				jsonx.PropertyPath("repositoryQuery"),
				c.RepositoryQuery,
				nil,
				jsonx.FormatOptions{InsertSpaces: true, TabSize: 2},
			)

			if err != nil {
				return errors.Wrapf(err, "%s.compute-property-edit", prefix)
			}

			edited, err := jsonx.ApplyEdits(svc.Config, edits...)
			if err != nil {
				return errors.Wrapf(err, "%s.apply-edits", prefix)
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
