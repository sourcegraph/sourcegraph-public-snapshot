package graphqlbackend

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
)

var extsvcConfigAllowEdits, _ = strconv.ParseBool(env.Get("EXTSVC_CONFIG_ALLOW_EDITS", "false", "When EXTSVC_CONFIG_FILE is in use, allow edits in the application to be made which will be overwritten on next process restart"))

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
	if os.Getenv("EXTSVC_CONFIG_FILE") != "" && !extsvcConfigAllowEdits {
		return nil, errors.New("adding external service not allowed when using EXTSVC_CONFIG_FILE")
	}

	externalService := &types.ExternalService{
		Kind:        args.Input.Kind,
		DisplayName: args.Input.DisplayName,
		Config:      args.Input.Config,
	}

	if err := db.ExternalServices.Create(ctx, conf.Get, externalService); err != nil {
		return nil, err
	}

	res := &externalServiceResolver{externalService: externalService}
	if err := syncExternalService(ctx, externalService); err != nil {
		res.warning = fmt.Sprintf("External service created, but we encountered a problem while validating the external service: %s", err)
	}

	return res, nil
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
	if os.Getenv("EXTSVC_CONFIG_FILE") != "" && !extsvcConfigAllowEdits {
		return nil, errors.New("updating external service not allowed when using EXTSVC_CONFIG_FILE")
	}

	if args.Input.Config != nil && strings.TrimSpace(*args.Input.Config) == "" {
		return nil, fmt.Errorf("blank external service configuration is invalid (must be valid JSONC)")
	}

	ps := conf.Get().Critical.AuthProviders
	update := &db.ExternalServiceUpdate{
		DisplayName: args.Input.DisplayName,
		Config:      args.Input.Config,
	}
	if err := db.ExternalServices.Update(ctx, ps, externalServiceID, update); err != nil {
		return nil, err
	}

	externalService, err := db.ExternalServices.GetByID(ctx, externalServiceID)
	if err != nil {
		return nil, err
	}

	res := &externalServiceResolver{externalService: externalService}
	if err = syncExternalService(ctx, externalService); err != nil {
		res.warning = fmt.Sprintf("External service updated, but we encountered a problem while validating the external service: %s", err)
	}

	return res, nil
}

// Eagerly trigger a repo-updater sync.
func syncExternalService(ctx context.Context, svc *types.ExternalService) error {
	_, err := repoupdater.DefaultClient.SyncExternalService(ctx, api.ExternalService{
		ID:          svc.ID,
		Kind:        svc.Kind,
		DisplayName: svc.DisplayName,
		Config:      svc.Config,
		CreatedAt:   svc.CreatedAt,
		UpdatedAt:   svc.UpdatedAt,
		DeletedAt:   svc.DeletedAt,
	})
	if err != nil {
		return err
	}

	return nil
}

func (*schemaResolver) DeleteExternalService(ctx context.Context, args *struct {
	ExternalService graphql.ID
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can delete external services.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	if os.Getenv("EXTSVC_CONFIG_FILE") != "" && !extsvcConfigAllowEdits {
		return nil, errors.New("deleting external service not allowed when using EXTSVC_CONFIG_FILE")
	}

	id, err := unmarshalExternalServiceID(args.ExternalService)
	if err != nil {
		return nil, err
	}

	externalService, err := db.ExternalServices.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := db.ExternalServices.Delete(ctx, id); err != nil {
		return nil, err
	}

	// The user doesn't care if triggering syncing failed when deleting a
	// service, so kick off in the background.
	go func() {
		_ = syncExternalService(context.Background(), externalService)
	}()

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

type computedExternalServiceConnectionResolver struct {
	args             graphqlutil.ConnectionArgs
	externalServices []*types.ExternalService
}

func (r *computedExternalServiceConnectionResolver) Nodes(ctx context.Context) []*externalServiceResolver {
	svcs := r.externalServices
	if r.args.First != nil && int(*r.args.First) < len(svcs) {
		svcs = svcs[:*r.args.First]
	}
	resolvers := make([]*externalServiceResolver, 0, len(svcs))
	for _, svc := range svcs {
		resolvers = append(resolvers, &externalServiceResolver{externalService: svc})
	}
	return resolvers
}

func (r *computedExternalServiceConnectionResolver) TotalCount(ctx context.Context) int32 {
	return int32(len(r.externalServices))
}

func (r *computedExternalServiceConnectionResolver) PageInfo(ctx context.Context) *graphqlutil.PageInfo {
	return graphqlutil.HasNextPage(r.args.First != nil && len(r.externalServices) >= int(*r.args.First))
}
