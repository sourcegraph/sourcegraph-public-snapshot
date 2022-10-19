package graphqlbackend

import (
	"context"
	"encoding/json"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *schemaResolver) SetExternalServiceRepos(ctx context.Context, args struct {
	ID       graphql.ID
	Repos    *[]string
	AllRepos bool
}) (*EmptyResponse, error) {
	start := time.Now()
	var err error
	defer reportExternalServiceDuration(start, SetRepos, &err)

	id, err := UnmarshalExternalServiceID(args.ID)
	if err != nil {
		return nil, err
	}

	extsvcStore := r.db.ExternalServices()
	es, err := extsvcStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: make sure the user has access to external service
	if err = backend.CheckExternalServiceAccess(ctx, r.db); err != nil {
		return nil, err
	}

	cfg, err := es.Configuration(ctx)
	if err != nil {
		return nil, err
	}

	ra, ok := cfg.(repoSetter)
	if !ok {
		err = errors.Errorf("ExternalService %s (kind %s) does not implement repoSetter", args.ID, es.Kind)
		return nil, err
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
	es.Config.Set(string(buf))

	// set to time.Zero to sync ASAP
	es.NextSyncAt = time.Time{}
	es.UpdatedAt = timeutil.Now()

	err = extsvcStore.Upsert(ctx, es)
	if err != nil {
		return nil, err
	}

	if err = backend.SyncExternalService(ctx, r.logger, es, 5*time.Second, r.repoupdaterClient); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type repoSetter interface {
	SetRepos(all bool, repos []string) error
}
