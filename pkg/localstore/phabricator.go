package localstore

import (
	"context"
	"fmt"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type phabricator struct{}

type errPhabricatorRepoNotFound struct {
	args []interface{}
}

func (err errPhabricatorRepoNotFound) Error() string {
	return fmt.Sprintf("phabricator repo not found: %v", err.args)
}

func (*phabricator) Create(ctx context.Context, callsign string, uri string) (*sourcegraph.PhabricatorRepo, error) {
	r := &sourcegraph.PhabricatorRepo{
		Callsign: callsign,
		URI:      uri,
	}
	err := globalDB.QueryRowContext(ctx,
		"INSERT INTO phabricator_repos(callsign, uri) VALUES($1, $2) RETURNING id",
		r.Callsign, r.URI).Scan(&r.ID)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (p *phabricator) CreateIfNotExists(ctx context.Context, callsign string, uri string) (*sourcegraph.PhabricatorRepo, error) {
	repo, err := p.getByCallsignOrURI(ctx, callsign, uri)
	if err != nil {
		if _, ok := err.(errPhabricatorRepoNotFound); !ok {
			return nil, err
		}
		return p.Create(ctx, callsign, uri)
	}
	return repo, nil
}

func (*phabricator) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.PhabricatorRepo, error) {
	rows, err := globalDB.QueryContext(ctx, "SELECT id, callsign, uri FROM phabricator_repos "+query, args...)
	if err != nil {
		return nil, err
	}

	repos := []*sourcegraph.PhabricatorRepo{}
	defer rows.Close()
	for rows.Next() {
		r := sourcegraph.PhabricatorRepo{}
		err := rows.Scan(&r.ID, &r.Callsign, &r.URI)
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

func (p *phabricator) getByCallsign(ctx context.Context, callsign string) (*sourcegraph.PhabricatorRepo, error) {
	return p.getOneBySQL(ctx, "WHERE callsign=$1", callsign)
}

func (p *phabricator) GetByURI(ctx context.Context, uri string) (*sourcegraph.PhabricatorRepo, error) {
	return p.getOneBySQL(ctx, "WHERE uri=$1", uri)
}

func (p *phabricator) getByCallsignOrURI(ctx context.Context, callsign string, uri string) (*sourcegraph.PhabricatorRepo, error) {
	return p.getOneBySQL(ctx, "WHERE callsign=$1 OR uri=$2", callsign, uri)
}
