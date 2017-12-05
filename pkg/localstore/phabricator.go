package localstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

type phabricator struct{}

type errPhabricatorRepoNotFound struct {
	args []interface{}
}

// DEPRECATED: use PHABRICATOR_CONFIG instead
// This environment variable determines the value to use to backfill an empty 'url' column.
var phabricatorURL = env.Get("PHABRICATOR_URL", "", "URL for internal Phabricator instance (on-prem)")

func (p *phabricator) BackfillURL() error {
	// If this exceeds the timeout (e.g., DB lock), there are probably other problems
	// occurring, but it will help debugging if we fail faster and log an error in that case.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repos, err := p.getBySQL(ctx, "WHERE url=$1", "")
	if err != nil {
		return err
	}

	if len(repos) != 0 && phabricatorURL == "" {
		return errors.New("cannot backfill phabricator_repos table without setting PHABRICATOR_URL environment")
	}

	for _, repo := range repos {
		_, err = globalDB.ExecContext(ctx, "UPDATE phabricator_repos SET url=$1 WHERE uri=$2", phabricatorURL, repo.URI)
		if err != nil {
			return err
		}
	}

	return nil
}

func (err errPhabricatorRepoNotFound) Error() string {
	return fmt.Sprintf("phabricator repo not found: %v", err.args)
}

func (*phabricator) Create(ctx context.Context, callsign string, uri string, phabURL string) (*sourcegraph.PhabricatorRepo, error) {
	r := &sourcegraph.PhabricatorRepo{
		Callsign: callsign,
		URI:      uri,
		URL:      phabURL,
	}
	err := globalDB.QueryRowContext(
		ctx,
		"INSERT INTO phabricator_repos(callsign, uri, url) VALUES($1, $2) RETURNING id",
		r.Callsign, r.URI, r.URL).Scan(&r.ID)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (p *phabricator) CreateIfNotExists(ctx context.Context, callsign string, uri string, phabURL string) (*sourcegraph.PhabricatorRepo, error) {
	repo, err := p.GetByURI(ctx, uri)
	if err != nil {
		if _, ok := err.(errPhabricatorRepoNotFound); !ok {
			return nil, err
		}
		return p.Create(ctx, callsign, uri, phabURL)
	}
	return repo, nil
}

func (*phabricator) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.PhabricatorRepo, error) {
	rows, err := globalDB.QueryContext(ctx, "SELECT id, callsign, uri, url FROM phabricator_repos "+query, args...)
	if err != nil {
		return nil, err
	}

	repos := []*sourcegraph.PhabricatorRepo{}
	defer rows.Close()
	for rows.Next() {
		r := sourcegraph.PhabricatorRepo{}
		err := rows.Scan(&r.ID, &r.Callsign, &r.URI, &r.URL)
		if err != nil {
			return nil, err
		}
		repos = append(repos, &r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return repos, nil
}

func (p *phabricator) getOneBySQL(ctx context.Context, query string, args ...interface{}) (*sourcegraph.PhabricatorRepo, error) {
	rows, err := p.getBySQL(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		return nil, errPhabricatorRepoNotFound{args}
	}
	return rows[0], nil
}

func (p *phabricator) GetByURI(ctx context.Context, uri string) (*sourcegraph.PhabricatorRepo, error) {
	return p.getOneBySQL(ctx, "WHERE uri=$1", uri)
}
