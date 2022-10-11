package graphqlbackend

import (
	"context"
	"encoding/json"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/conf"
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
	var namespaceUserID, namespaceOrgID int32
	defer reportExternalServiceDuration(start, SetRepos, &err, &namespaceUserID, &namespaceOrgID)

	id, err := UnmarshalExternalServiceID(args.ID)
	if err != nil {
		return nil, err
	}

	extsvcStore := r.db.ExternalServices()
	es, err := extsvcStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	namespaceUserID, namespaceOrgID = es.NamespaceUserID, es.NamespaceOrgID

	// 🚨 SECURITY: make sure the user has access to external service
	if err = backend.CheckExternalServiceAccess(ctx, r.db, es.NamespaceUserID, es.NamespaceOrgID); err != nil {
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
		// If we know the number of repos up front we can ensure that they don't exceed
		// their limit before hitting the code host
		maxAllowed := conf.UserReposMaxPerUser()
		if (es.NamespaceUserID != 0 || es.NamespaceOrgID != 0) && len(repos) > maxAllowed {
			err = errors.Errorf("Too many repositories, %d. Sourcegraph supports adding a maximum of %d repositories.", len(repos), maxAllowed)
			return nil, err
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
