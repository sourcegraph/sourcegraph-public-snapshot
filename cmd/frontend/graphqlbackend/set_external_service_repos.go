package graphqlbackend

import (
	"context"
	"encoding/json"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/db"
)

func (r *schemaResolver) SetExternalServiceRepos(ctx context.Context, args struct {
	ID       graphql.ID
	Repos    *[]string
	AllRepos bool
}) (*EmptyResponse, error) {
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

	ra, ok := cfg.(repoSetter)
	if !ok {
		return nil, errors.Errorf("ExternalService %s (kind %s) does not implement repoSetter", args.ID, es.Kind)
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
	if err != nil {
		return nil, err
	}
	es.Config = string(buf)

	// set to time.Zero to sync ASAP
	es.NextSyncAt = time.Time{}

	err = db.ExternalServices.Upsert(ctx, es)
	if err != nil {
		return nil, err
	}

	if err := syncExternalService(ctx, es); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type repoSetter interface {
	SetRepos(all bool, repos []string) error
}
