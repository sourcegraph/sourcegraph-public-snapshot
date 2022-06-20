package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type PhabricatorStore interface {
	Create(ctx context.Context, callsign string, name api.RepoName, phabURL string) (*types.PhabricatorRepo, error)
	CreateIfNotExists(ctx context.Context, callsign string, name api.RepoName, phabURL string) (*types.PhabricatorRepo, error)
	CreateOrUpdate(ctx context.Context, callsign string, name api.RepoName, phabURL string) (*types.PhabricatorRepo, error)
	GetByName(context.Context, api.RepoName) (*types.PhabricatorRepo, error)
	Transact(context.Context) (PhabricatorStore, error)
	With(basestore.ShareableStore) PhabricatorStore
	basestore.ShareableStore
}

type phabricatorStore struct {
	*basestore.Store
}

// PhabricatorWith instantiates and returns a new PhabricatorStore using the other store handle.
func PhabricatorWith(other basestore.ShareableStore) PhabricatorStore {
	return &phabricatorStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *phabricatorStore) With(other basestore.ShareableStore) PhabricatorStore {
	return &phabricatorStore{Store: s.Store.With(other)}
}

func (s *phabricatorStore) Transact(ctx context.Context) (PhabricatorStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &phabricatorStore{Store: txBase}, err
}

type errPhabricatorRepoNotFound struct {
	args []any
}

func (err errPhabricatorRepoNotFound) Error() string {
	return fmt.Sprintf("phabricator repo not found: %v", err.args)
}

func (err errPhabricatorRepoNotFound) NotFound() bool { return true }

func (p *phabricatorStore) Create(ctx context.Context, callsign string, name api.RepoName, phabURL string) (*types.PhabricatorRepo, error) {
	r := &types.PhabricatorRepo{
		Callsign: callsign,
		Name:     name,
		URL:      phabURL,
	}
	err := p.Handle().QueryRowContext(
		ctx,
		"INSERT INTO phabricator_repos(callsign, repo_name, url) VALUES($1, $2, $3) RETURNING id",
		r.Callsign, r.Name, r.URL).Scan(&r.ID)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (p *phabricatorStore) CreateOrUpdate(ctx context.Context, callsign string, name api.RepoName, phabURL string) (*types.PhabricatorRepo, error) {
	r := &types.PhabricatorRepo{
		Callsign: callsign,
		Name:     name,
		URL:      phabURL,
	}
	err := p.Handle().QueryRowContext(
		ctx,
		"UPDATE phabricator_repos SET callsign=$1, url=$2, updated_at=now() WHERE repo_name=$3 RETURNING id",
		r.Callsign, r.URL, r.Name).Scan(&r.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return p.Create(ctx, callsign, name, phabURL)
		}
		return nil, err
	}
	return r, nil
}

func (p *phabricatorStore) CreateIfNotExists(ctx context.Context, callsign string, name api.RepoName, phabURL string) (*types.PhabricatorRepo, error) {
	repo, err := p.GetByName(ctx, name)
	if err != nil {
		if !errors.HasType(err, errPhabricatorRepoNotFound{}) {
			return nil, err
		}
		return p.Create(ctx, callsign, name, phabURL)
	}
	return repo, nil
}

func (p *phabricatorStore) getBySQL(ctx context.Context, query string, args ...any) ([]*types.PhabricatorRepo, error) {
	rows, err := p.Handle().QueryContext(ctx, "SELECT id, callsign, repo_name, url FROM phabricator_repos "+query, args...)
	if err != nil {
		return nil, err
	}

	repos := []*types.PhabricatorRepo{}
	defer rows.Close()
	for rows.Next() {
		r := types.PhabricatorRepo{}
		err := rows.Scan(&r.ID, &r.Callsign, &r.Name, &r.URL)
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

func (p *phabricatorStore) getOneBySQL(ctx context.Context, query string, args ...any) (*types.PhabricatorRepo, error) {
	rows, err := p.getBySQL(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		return nil, errPhabricatorRepoNotFound{args}
	}
	return rows[0], nil
}

func (p *phabricatorStore) GetByName(ctx context.Context, name api.RepoName) (*types.PhabricatorRepo, error) {
	opt := ExternalServicesListOptions{
		Kinds: []string{extsvc.KindPhabricator},
		LimitOffset: &LimitOffset{
			Limit: 500, // The number is randomly chosen
		},
	}
	for {
		svcs, err := ExternalServicesWith(p).List(ctx, opt)
		if err != nil {
			return nil, errors.Wrap(err, "list")
		}
		if len(svcs) == 0 {
			break // No more results, exiting
		}
		opt.AfterID = svcs[len(svcs)-1].ID // Advance the cursor

		for _, svc := range svcs {
			cfg, err := extsvc.ParseConfig(svc.Kind, svc.Config)
			if err != nil {
				return nil, errors.Wrap(err, "parse config")
			}

			var conn *schema.PhabricatorConnection
			switch c := cfg.(type) {
			case *schema.PhabricatorConnection:
				conn = c
			default:
				log15.Error("phabricator.GetByName", "error", errors.Errorf("want *schema.PhabricatorConnection but got %T", cfg))
				continue
			}

			for _, repo := range conn.Repos {
				if api.RepoName(repo.Path) == name {
					return &types.PhabricatorRepo{
						Name:     api.RepoName(repo.Path),
						Callsign: repo.Callsign,
						URL:      conn.Url,
					}, nil
				}
			}
		}

		if len(svcs) < opt.Limit {
			break // Less results than limit means we've reached end
		}
	}

	return p.getOneBySQL(ctx, "WHERE repo_name=$1", name)
}
