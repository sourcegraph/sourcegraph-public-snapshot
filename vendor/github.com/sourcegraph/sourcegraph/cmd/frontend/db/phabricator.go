package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/dbconn"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/atomicvalue"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

type phabricator struct{}

type errPhabricatorRepoNotFound struct {
	args []interface{}
}

var phabricatorRepos = atomicvalue.New()

func init() {
	conf.Watch(func() {
		phabricatorRepos.Set(func() interface{} {
			repos := map[api.RepoURI]*types.PhabricatorRepo{}
			for _, config := range conf.Get().Phabricator {
				for _, repo := range config.Repos {
					repos[api.RepoURI(repo.Path)] = &types.PhabricatorRepo{
						URI:      api.RepoURI(repo.Path),
						Callsign: repo.Callsign,
						URL:      config.Url,
					}
				}
			}
			return repos
		})
	})
}

func (err errPhabricatorRepoNotFound) Error() string {
	return fmt.Sprintf("phabricator repo not found: %v", err.args)
}

func (err errPhabricatorRepoNotFound) NotFound() bool { return true }

func (*phabricator) Create(ctx context.Context, callsign string, uri api.RepoURI, phabURL string) (*types.PhabricatorRepo, error) {
	r := &types.PhabricatorRepo{
		Callsign: callsign,
		URI:      uri,
		URL:      phabURL,
	}
	err := dbconn.Global.QueryRowContext(
		ctx,
		"INSERT INTO phabricator_repos(callsign, uri, url) VALUES($1, $2, $3) RETURNING id",
		r.Callsign, r.URI, r.URL).Scan(&r.ID)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (p *phabricator) CreateOrUpdate(ctx context.Context, callsign string, uri api.RepoURI, phabURL string) (*types.PhabricatorRepo, error) {
	r := &types.PhabricatorRepo{
		Callsign: callsign,
		URI:      uri,
		URL:      phabURL,
	}
	err := dbconn.Global.QueryRowContext(
		ctx,
		"UPDATE phabricator_repos SET callsign=$1, url=$2, updated_at=now() WHERE uri=$3 RETURNING id",
		r.Callsign, r.URL, r.URI).Scan(&r.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return p.Create(ctx, callsign, uri, phabURL)
		}
		return nil, err
	}
	return r, nil
}

func (p *phabricator) CreateIfNotExists(ctx context.Context, callsign string, uri api.RepoURI, phabURL string) (*types.PhabricatorRepo, error) {
	repo, err := p.GetByURI(ctx, uri)
	if err != nil {
		if _, ok := err.(errPhabricatorRepoNotFound); !ok {
			return nil, err
		}
		return p.Create(ctx, callsign, uri, phabURL)
	}
	return repo, nil
}

func (*phabricator) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*types.PhabricatorRepo, error) {
	rows, err := dbconn.Global.QueryContext(ctx, "SELECT id, callsign, uri, url FROM phabricator_repos "+query, args...)
	if err != nil {
		return nil, err
	}

	repos := []*types.PhabricatorRepo{}
	defer rows.Close()
	for rows.Next() {
		r := types.PhabricatorRepo{}
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

func (p *phabricator) getOneBySQL(ctx context.Context, query string, args ...interface{}) (*types.PhabricatorRepo, error) {
	rows, err := p.getBySQL(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		return nil, errPhabricatorRepoNotFound{args}
	}
	return rows[0], nil
}

func (p *phabricator) GetByURI(ctx context.Context, uri api.RepoURI) (*types.PhabricatorRepo, error) {
	if Mocks.Phabricator.GetByURI != nil {
		return Mocks.Phabricator.GetByURI(uri)
	}
	phabricatorRepos := phabricatorRepos.Get().(map[api.RepoURI]*types.PhabricatorRepo)
	if r := phabricatorRepos[uri]; r != nil {
		return r, nil
	}
	return p.getOneBySQL(ctx, "WHERE uri=$1", uri)
}

type MockPhabricator struct {
	GetByURI func(repo api.RepoURI) (*types.PhabricatorRepo, error)
}
