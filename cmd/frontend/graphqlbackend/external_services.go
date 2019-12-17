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

func (r *schemaResolver) AddCodeHost(ctx context.Context, args *struct {
	Input *struct {
		Kind        string
		DisplayName string
		Config      string
	}
}) (*codeHostResolver, error) {
	// 🚨 SECURITY: Only site admins may add external services.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	if os.Getenv("EXTSVC_CONFIG_FILE") != "" && !extsvcConfigAllowEdits {
		return nil, errors.New("adding external service not allowed when using EXTSVC_CONFIG_FILE")
	}

	codeHost := &types.CodeHost{
		Kind:        args.Input.Kind,
		DisplayName: args.Input.DisplayName,
		Config:      args.Input.Config,
	}

	if err := db.CodeHosts.Create(ctx, conf.Get, codeHost); err != nil {
		return nil, err
	}

	res := &codeHostResolver{codeHost: codeHost}
	if err := syncCodeHost(ctx, codeHost); err != nil {
		res.warning = fmt.Sprintf("External service created, but we encountered a problem while validating the external service: %s", err)
	}

	return res, nil
}

func (*schemaResolver) UpdateCodeHost(ctx context.Context, args *struct {
	Input *struct {
		ID          graphql.ID
		DisplayName *string
		Config      *string
	}
}) (*codeHostResolver, error) {
	codeHostID, err := unmarshalCodeHostID(args.Input.ID)
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

	ps := conf.Get().AuthProviders
	update := &db.CodeHostUpdate{
		DisplayName: args.Input.DisplayName,
		Config:      args.Input.Config,
	}
	if err := db.CodeHosts.Update(ctx, ps, codeHostID, update); err != nil {
		return nil, err
	}

	codeHost, err := db.CodeHosts.GetByID(ctx, codeHostID)
	if err != nil {
		return nil, err
	}

	res := &codeHostResolver{codeHost: codeHost}
	if err = syncCodeHost(ctx, codeHost); err != nil {
		res.warning = fmt.Sprintf("External service updated, but we encountered a problem while validating the external service: %s", err)
	}

	return res, nil
}

// Eagerly trigger a repo-updater sync.
func syncCodeHost(ctx context.Context, svc *types.CodeHost) error {
	_, err := repoupdater.DefaultClient.SyncCodeHost(ctx, api.CodeHost{
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

func (*schemaResolver) DeleteCodeHost(ctx context.Context, args *struct {
	CodeHost graphql.ID
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins can delete external services.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	if os.Getenv("EXTSVC_CONFIG_FILE") != "" && !extsvcConfigAllowEdits {
		return nil, errors.New("deleting external service not allowed when using EXTSVC_CONFIG_FILE")
	}

	id, err := unmarshalCodeHostID(args.CodeHost)
	if err != nil {
		return nil, err
	}

	codeHost, err := db.CodeHosts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := db.CodeHosts.Delete(ctx, id); err != nil {
		return nil, err
	}

	// The user doesn't care if triggering syncing failed when deleting a
	// service, so kick off in the background.
	go func() {
		_ = syncCodeHost(context.Background(), codeHost)
	}()

	return &EmptyResponse{}, nil
}

func (r *schemaResolver) CodeHosts(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
}) (*codeHostConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins may read external services (they have secrets).
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	var opt db.CodeHostsListOptions
	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &codeHostConnectionResolver{opt: opt}, nil
}

type codeHostConnectionResolver struct {
	opt db.CodeHostsListOptions

	// cache results because they are used by multiple fields
	once             sync.Once
	codeHosts []*types.CodeHost
	err              error
}

func (r *codeHostConnectionResolver) compute(ctx context.Context) ([]*types.CodeHost, error) {
	r.once.Do(func() {
		r.codeHosts, r.err = db.CodeHosts.List(ctx, r.opt)
	})
	return r.codeHosts, r.err
}

func (r *codeHostConnectionResolver) Nodes(ctx context.Context) ([]*codeHostResolver, error) {
	codeHosts, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*codeHostResolver, 0, len(codeHosts))
	for _, codeHost := range codeHosts {
		resolvers = append(resolvers, &codeHostResolver{codeHost: codeHost})
	}
	return resolvers, nil
}

func (r *codeHostConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := db.CodeHosts.Count(ctx, r.opt)
	return int32(count), err
}

func (r *codeHostConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	codeHosts, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && len(codeHosts) >= r.opt.Limit), nil
}

type computedCodeHostConnectionResolver struct {
	args             graphqlutil.ConnectionArgs
	codeHosts []*types.CodeHost
}

func (r *computedCodeHostConnectionResolver) Nodes(ctx context.Context) []*codeHostResolver {
	svcs := r.codeHosts
	if r.args.First != nil && int(*r.args.First) < len(svcs) {
		svcs = svcs[:*r.args.First]
	}
	resolvers := make([]*codeHostResolver, 0, len(svcs))
	for _, svc := range svcs {
		resolvers = append(resolvers, &codeHostResolver{codeHost: svc})
	}
	return resolvers
}

func (r *computedCodeHostConnectionResolver) TotalCount(ctx context.Context) int32 {
	return int32(len(r.codeHosts))
}

func (r *computedCodeHostConnectionResolver) PageInfo(ctx context.Context) *graphqlutil.PageInfo {
	return graphqlutil.HasNextPage(r.args.First != nil && len(r.codeHosts) >= int(*r.args.First))
}
