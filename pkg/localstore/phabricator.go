package localstore

import (
	"context"
	"fmt"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

type phabricator struct{}

type errPhabricatorRepoNotFound struct {
	args []interface{}
}

var (
	phabricatorConfigs = conf.Get().Phabricator
	phabricatorRepos   map[string]*sourcegraph.PhabricatorRepo
)

func init() {
	phabricatorRepos = map[string]*sourcegraph.PhabricatorRepo{}

	for _, config := range phabricatorConfigs {
		for _, repo := range config.Repos {
			phabricatorRepos[repo.Path] = &sourcegraph.PhabricatorRepo{
				URI:      repo.Path,
				Callsign: repo.Callsign,
				URL:      config.Url,
			}
		}
	}
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
		"INSERT INTO phabricator_repos(callsign, uri, url) VALUES($1, $2, $3) RETURNING id",
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
	if r := phabricatorRepos[uri]; r != nil {
		return r, nil
	}
	return p.getOneBySQL(ctx, "WHERE uri=$1", uri)
}
