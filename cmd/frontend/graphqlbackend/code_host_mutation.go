package graphqlbackend

import (
	"context"
	"encoding/json"
	"github.com/graph-gophers/graphql-go"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/db"
)

func (r *schemaResolver) SetCodeHostRepos(ctx context.Context, args struct {
	ID       graphql.ID
	Repos    *[]string
	AllRepos bool
}) (*EmptyResponse, error) {
	log15.Warn("SETTING REPOS")
	id, err := unmarshalExternalServiceID(args.ID)
	if err != nil {
		return nil, err
	}

	es, err := db.ExternalServices.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: make sure the user is either site admin or the same user being requested
	if err := backend.CheckSiteAdminOrSameUser(ctx, es.NamespaceUserID); err != nil {
		return nil, err
	}

	cfg, err := es.Configuration()
	if err != nil {
		return nil, err
	}

	ra, ok := cfg.(extsvc.ReposSetter)
	if !ok {
		return nil, errors.Errorf("Code host %s (kind %s) does not implement extsvc.RepoAdder", args.ID, es.Kind)
	}

	var repos []string
	if args.Repos != nil {
		repos = *args.Repos
	}
	err = ra.SetRepos(args.AllRepos, repos)
	if err != nil {
		return nil, err
	}

	buf, err := json.MarshalIndent(cfg, "", "  ")
	es.Config = string(buf)

	err = db.ExternalServices.Upsert(ctx, es)
	if err != nil {
		return nil, err
	}

	if err := syncExternalService(ctx, es); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}
