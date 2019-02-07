package graphqlbackend

import (
	"context"
	"fmt"
	"strings"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
)

func (r *schemaResolver) AddExternalService(ctx context.Context, args *struct {
	Input *struct {
		Kind        string
		DisplayName string
		Config      string
	}
}) (*externalServiceResolver, error) {
	// 🚨 SECURITY: Only site admins may add external services.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	externalService := &types.ExternalService{
		Kind:        args.Input.Kind,
		DisplayName: args.Input.DisplayName,
		Config:      args.Input.Config,
	}

	if err := db.ExternalServices.Create(ctx, externalService); err != nil {
		return nil, err
	}

	if err := syncExternalService(ctx, externalService); err != nil {
		return nil, errors.Wrap(err, "external service created, but sync request failed")
	}

	return &externalServiceResolver{externalService: externalService}, nil
}

func (*schemaResolver) UpdateExternalService(ctx context.Context, args *struct {
	Input *struct {
		ID          graphql.ID
		DisplayName *string
		Config      *string
	}
}) (*externalServiceResolver, error) {
	externalServiceID, err := unmarshalExternalServiceID(args.Input.ID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only site admins are allowed to update the user.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	if args.Input.Config != nil && strings.TrimSpace(*args.Input.Config) == "" {
		return nil, fmt.Errorf("blank external service configuration is invalid (must be valid JSONC)")
	}

	update := &db.ExternalServiceUpdate{
		DisplayName: args.Input.DisplayName,
		Config:      args.Input.Config,
	}
	if err := db.ExternalServices.Update(ctx, externalServiceID, update); err != nil {
		return nil, err
	}

	externalService, err := db.ExternalServices.GetByID(ctx, externalServiceID)
	if err != nil {
		return nil, err
	}

	if err = syncExternalService(ctx, externalService); err != nil {
		return nil, errors.Wrap(err, "external service updated, but sync request failed")
	}

	return &externalServiceResolver{externalService: externalService}, nil
}

// Eagerly trigger a repo-updater sync.
func syncExternalService(ctx context.Context, svc *types.ExternalService) error {
	return repoupdater.DefaultClient.SyncExternalService(ctx, api.ExternalService{
		ID:          svc.ID,
		Kind:        svc.Kind,
		DisplayName: svc.DisplayName,
		Config:      svc.Config,
		CreatedAt:   svc.CreatedAt,
		UpdatedAt:   svc.UpdatedAt,
		DeletedAt:   svc.DeletedAt,
	})
}

func (*schemaResolver) DeleteExternalService(ctx context.Context, args *struct {
	ExternalService graphql.ID
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can delete external services.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	id, err := unmarshalExternalServiceID(args.ExternalService)
	if err != nil {
		return nil, err
	}

	if err := db.ExternalServices.Delete(ctx, id); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

func (r *schemaResolver) ExternalServices(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
}) (*externalServiceConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins may read external services (they have secrets).
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	var opt db.ExternalServicesListOptions
	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &externalServiceConnectionResolver{opt: opt}, nil
}

type externalServiceConnectionResolver struct {
	opt db.ExternalServicesListOptions

	// cache results because they are used by multiple fields
	once             sync.Once
	externalServices []*types.ExternalService
	err              error
}

func (r *externalServiceConnectionResolver) compute(ctx context.Context) ([]*types.ExternalService, error) {
	r.once.Do(func() {
		r.externalServices, r.err = db.ExternalServices.List(ctx, r.opt)
	})
	return r.externalServices, r.err
}

func (r *externalServiceConnectionResolver) Nodes(ctx context.Context) ([]*externalServiceResolver, error) {
	externalServices, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*externalServiceResolver, 0, len(externalServices))
	for _, externalService := range externalServices {
		resolvers = append(resolvers, &externalServiceResolver{externalService: externalService})
	}
	return resolvers, nil
}

func (r *externalServiceConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := db.ExternalServices.Count(ctx, r.opt)
	return int32(count), err
}

func (r *externalServiceConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	externalServices, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && len(externalServices) >= r.opt.Limit), nil
}
