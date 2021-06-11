package graphqlbackend

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var extsvcConfigAllowEdits, _ = strconv.ParseBool(env.Get("EXTSVC_CONFIG_ALLOW_EDITS", "false", "When EXTSVC_CONFIG_FILE is in use, allow edits in the application to be made which will be overwritten on next process restart"))

type addExternalServiceArgs struct {
	Input addExternalServiceInput
}

type addExternalServiceInput struct {
	Kind        string
	DisplayName string
	Config      string
	Namespace   *graphql.ID
}

func (r *schemaResolver) AddExternalService(ctx context.Context, args *addExternalServiceArgs) (*externalServiceResolver, error) {
	if os.Getenv("EXTSVC_CONFIG_FILE") != "" && !extsvcConfigAllowEdits {
		return nil, errors.New("adding external service not allowed when using EXTSVC_CONFIG_FILE")
	}

	// 🚨 SECURITY: Only site admins may add external services if user mode is disabled.
	namespaceUserID := int32(0)
	isSiteAdmin := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db) == nil
	allowUserExternalServices, err := database.Users(r.db).CurrentUserAllowedExternalServices(ctx)
	if err != nil {
		return nil, err
	}

	if args.Input.Namespace != nil {
		if allowUserExternalServices == conf.ExternalServiceModeDisabled {
			return nil, errors.New("allow users to add external services is not enabled")
		}

		var err error
		switch relay.UnmarshalKind(*args.Input.Namespace) {
		case "User":
			err = relay.UnmarshalSpec(*args.Input.Namespace, &namespaceUserID)
		default:
			err = errors.Errorf("invalid namespace %q", *args.Input.Namespace)
		}

		if err != nil {
			return nil, err
		}

		if namespaceUserID != actor.FromContext(ctx).UID {
			return nil, errors.New("the namespace is not same as the authenticated user")
		}

	} else if !isSiteAdmin {
		return nil, backend.ErrMustBeSiteAdmin
	}

	externalService := &types.ExternalService{
		Kind:        args.Input.Kind,
		DisplayName: args.Input.DisplayName,
		Config:      args.Input.Config,
	}
	if namespaceUserID > 0 {
		externalService.NamespaceUserID = namespaceUserID
	}

	if err := database.ExternalServices(r.db).Create(ctx, conf.Get, externalService); err != nil {
		return nil, err
	}

	res := &externalServiceResolver{db: r.db, externalService: externalService}
	if err := syncExternalService(ctx, externalService, 5*time.Second, r.repoupdaterClient); err != nil {
		res.warning = fmt.Sprintf("External service created, but we encountered a problem while validating the external service: %s", err)
	}

	return res, nil
}

type updateExternalServiceArgs struct {
	Input updateExternalServiceInput
}

type updateExternalServiceInput struct {
	ID          graphql.ID
	DisplayName *string
	Config      *string
}

func (r *schemaResolver) UpdateExternalService(ctx context.Context, args *updateExternalServiceArgs) (*externalServiceResolver, error) {
	if os.Getenv("EXTSVC_CONFIG_FILE") != "" && !extsvcConfigAllowEdits {
		return nil, errors.New("updating external service not allowed when using EXTSVC_CONFIG_FILE")
	}

	id, err := unmarshalExternalServiceID(args.Input.ID)
	if err != nil {
		return nil, err
	}

	es, err := database.ExternalServices(r.db).GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Site admins can only update site level external services.
	// Otherwise, the current user can only update their own external services.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		if es.NamespaceUserID == 0 {
			return nil, err
		} else if actor.FromContext(ctx).UID != es.NamespaceUserID {
			return nil, errNoAccessExternalService
		}
	}

	if args.Input.Config != nil && strings.TrimSpace(*args.Input.Config) == "" {
		return nil, errors.New("blank external service configuration is invalid (must be valid JSONC)")
	}

	ps := conf.Get().AuthProviders
	update := &database.ExternalServiceUpdate{
		DisplayName: args.Input.DisplayName,
		Config:      args.Input.Config,
	}
	if err := database.ExternalServices(r.db).Update(ctx, ps, id, update); err != nil {
		return nil, err
	}

	// Fetch from database again to get all fields with updated values.
	es, err = database.ExternalServices(r.db).GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	res := &externalServiceResolver{db: r.db, externalService: es}
	if err = syncExternalService(ctx, es, 5*time.Second, r.repoupdaterClient); err != nil {
		res.warning = fmt.Sprintf("External service updated, but we encountered a problem while validating the external service: %s", err)
	}

	return res, nil
}

// repoupdaterClient is an interface with only the methods required in syncExternalService. As a
// result instead of using the entire repoupdater client implementation, we use a thinner API which
// only needs the SyncExternalService method to be defined on the object.
type repoupdaterClient interface {
	SyncExternalService(ctx context.Context, svc api.ExternalService) (*protocol.ExternalServiceSyncResult, error)
}

// syncExternalService will eagerly trigger a repo-updater sync. It accepts a
// timeout as an argument which is recommended to be 5 seconds unless the caller
// has special requirements for it to be larger or smaller.
func syncExternalService(ctx context.Context, svc *types.ExternalService, timeout time.Duration, client repoupdaterClient) error {
	// Set a timeout to validate external service sync. It usually fails in
	// under 5s if there is a problem.
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := client.SyncExternalService(ctx, api.ExternalService{
		ID:              svc.ID,
		Kind:            svc.Kind,
		DisplayName:     svc.DisplayName,
		Config:          svc.Config,
		CreatedAt:       svc.CreatedAt,
		UpdatedAt:       svc.UpdatedAt,
		DeletedAt:       svc.DeletedAt,
		LastSyncAt:      svc.LastSyncAt,
		NextSyncAt:      svc.NextSyncAt,
		NamespaceUserID: svc.NamespaceUserID,
	})

	// If context error is anything but a deadline exceeded error, we do not want to propagate
	// it. But we definitely want to log the error as a warning.
	if ctx.Err() != nil && ctx.Err() != context.DeadlineExceeded {
		log15.Warn("syncExternalService: context error discarded", "err", ctx.Err())
		return nil
	}

	// err is either nil or contains an actual error from the API call. And we return it
	// nonetheless.
	return errors.Wrapf(err, "error in syncExternalService for service %q with ID %d", svc.Kind, svc.ID)
}

type deleteExternalServiceArgs struct {
	ExternalService graphql.ID
}

func (r *schemaResolver) DeleteExternalService(ctx context.Context, args *deleteExternalServiceArgs) (*EmptyResponse, error) {
	if os.Getenv("EXTSVC_CONFIG_FILE") != "" && !extsvcConfigAllowEdits {
		return nil, errors.New("deleting external service not allowed when using EXTSVC_CONFIG_FILE")
	}

	id, err := unmarshalExternalServiceID(args.ExternalService)
	if err != nil {
		return nil, err
	}

	es, err := database.ExternalServices(r.db).GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only site admins may delete all or a user's external services.
	// Otherwise, the authenticated user can only delete external services under the same namespace.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		if es.NamespaceUserID == 0 {
			return nil, err
		} else if actor.FromContext(ctx).UID != es.NamespaceUserID {
			return nil, errNoAccessExternalService
		}
	}

	if err := database.ExternalServices(r.db).Delete(ctx, id); err != nil {
		return nil, err
	}
	now := time.Now()
	es.DeletedAt = now

	// The user doesn't care if triggering syncing failed when deleting a
	// service, so kick off in the background.
	go func() {
		if err := syncExternalService(context.Background(), es, 5*time.Second, r.repoupdaterClient); err != nil {
			log15.Warn("Performing final sync after external service deletion", "err", err)
		}
	}()

	return &EmptyResponse{}, nil
}

type ExternalServicesArgs struct {
	Namespace *graphql.ID
	graphqlutil.ConnectionArgs
	After *string
}

var errNoAccessExternalService = errors.New("the authenticated user does not have access to this external service")

// checkExternalServiceAccess checks whether the current user is allowed to
// access the supplied external service.
//
// 🚨 SECURITY: Site admins can view external services with no owner, otherwise
// only the owner of the external service is allowed to access it.
func checkExternalServiceAccess(ctx context.Context, db dbutil.DB, namespaceUserID int32) error {
	// Fast path that doesn't need to hit DB as we can get id from context
	if a := actor.FromContext(ctx); a.IsAuthenticated() && namespaceUserID == a.UID {
		return nil
	}

	// Special case when external service has no owner
	if namespaceUserID == 0 && backend.CheckCurrentUserIsSiteAdmin(ctx, db) == nil {
		return nil
	}

	return errNoAccessExternalService
}

func (r *schemaResolver) ExternalServices(ctx context.Context, args *ExternalServicesArgs) (*externalServiceConnectionResolver, error) {
	var namespaceUserID int32
	if args.Namespace != nil {
		var err error
		switch relay.UnmarshalKind(*args.Namespace) {
		case "User":
			err = relay.UnmarshalSpec(*args.Namespace, &namespaceUserID)
		default:
			err = errors.Errorf("invalid namespace %q", *args.Namespace)
		}

		if err != nil {
			return nil, err
		}
	}

	if err := checkExternalServiceAccess(ctx, r.db, namespaceUserID); err != nil {
		return nil, err
	}

	var afterID int64
	if args.After != nil {
		var err error
		afterID, err = unmarshalExternalServiceID(graphql.ID(*args.After))
		if err != nil {
			return nil, err
		}
	}

	opt := database.ExternalServicesListOptions{
		NamespaceUserID: namespaceUserID,
		AfterID:         afterID,
	}
	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &externalServiceConnectionResolver{db: r.db, opt: opt}, nil
}

type externalServiceConnectionResolver struct {
	opt database.ExternalServicesListOptions

	// cache results because they are used by multiple fields
	once             sync.Once
	externalServices []*types.ExternalService
	err              error
	db               dbutil.DB
}

func (r *externalServiceConnectionResolver) compute(ctx context.Context) ([]*types.ExternalService, error) {
	r.once.Do(func() {
		r.externalServices, r.err = database.ExternalServices(r.db).List(ctx, r.opt)
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
		resolvers = append(resolvers, &externalServiceResolver{db: r.db, externalService: externalService})
	}
	return resolvers, nil
}

func (r *externalServiceConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	// Reset pagination cursor to get correct total count
	opt := r.opt
	opt.AfterID = 0
	count, err := database.ExternalServices(r.db).Count(ctx, opt)
	return int32(count), err
}

func (r *externalServiceConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	externalServices, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	// We would have had all results when no limit set
	if r.opt.LimitOffset == nil {
		return graphqlutil.HasNextPage(false), nil
	}

	// We got less results than limit, means we've had all results
	if len(externalServices) < r.opt.Limit {
		return graphqlutil.HasNextPage(false), nil
	}

	// In case the number of results happens to be the same as the limit,
	// we need another query to get accurate total count with same cursor
	// to determine if there are more results than the limit we set.
	count, err := database.ExternalServices(r.db).Count(ctx, r.opt)
	if err != nil {
		return nil, err
	}

	if count > len(externalServices) {
		endCursorID := externalServices[len(externalServices)-1].ID
		return graphqlutil.NextPageCursor(string(marshalExternalServiceID(endCursorID))), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

type computedExternalServiceConnectionResolver struct {
	args             graphqlutil.ConnectionArgs
	externalServices []*types.ExternalService
	db               dbutil.DB
}

func (r *computedExternalServiceConnectionResolver) Nodes(ctx context.Context) []*externalServiceResolver {
	svcs := r.externalServices
	if r.args.First != nil && int(*r.args.First) < len(svcs) {
		svcs = svcs[:*r.args.First]
	}
	resolvers := make([]*externalServiceResolver, 0, len(svcs))
	for _, svc := range svcs {
		resolvers = append(resolvers, &externalServiceResolver{db: r.db, externalService: svc})
	}
	return resolvers
}

func (r *computedExternalServiceConnectionResolver) TotalCount(ctx context.Context) int32 {
	return int32(len(r.externalServices))
}

func (r *computedExternalServiceConnectionResolver) PageInfo(ctx context.Context) *graphqlutil.PageInfo {
	return graphqlutil.HasNextPage(r.args.First != nil && len(r.externalServices) >= int(*r.args.First))
}
