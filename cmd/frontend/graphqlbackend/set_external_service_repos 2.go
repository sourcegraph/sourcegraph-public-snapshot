package graphqlbackend

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
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

	extsvcStore := database.ExternalServices(r.db)
	es, err := extsvcStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: make sure the user is either site admin or the same user being requested
	if err := backend.CheckSiteAdminOrSameUser(ctx, r.db, es.NamespaceUserID); err != nil {
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
		// If we know the number of repos up front we can ensure that they don't exceed
		// their limit before hitting the code host
		maxAllowed := conf.UserReposMaxPerUser()
		if es.NamespaceUserID != 0 && len(repos) > maxAllowed {
			return nil, errors.Errorf("Too many repositories, %d. Sourcegraph supports adding a maximum of %d repositories.", len(repos), maxAllowed)
		}
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
	es.UpdatedAt = timeutil.Now()

	err = extsvcStore.Upsert(ctx, es)
	if err != nil {
		return nil, err
	}

	if err := syncExternalService(ctx, es, 5*time.Second, r.repoupdaterClient); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type repoSetter interface {
	SetRepos(all bool, repos []string) error
}
