package codyaccessservice

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/connectutil"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/codyaccess"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/samsm2m"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	codyaccessv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/codyaccess/v1"
	codyaccessv1connect "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/codyaccess/v1/v1connect"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

const Name = codyaccessv1connect.CodyAccessServiceName

func RegisterV1(
	logger log.Logger,
	mux *http.ServeMux,
	store StoreV1,
	opts ...connect.HandlerOption,
) {
	mux.Handle(
		codyaccessv1connect.NewCodyAccessServiceHandler(
			NewHandlerV1(logger, store),
			opts...,
		),
	)
}

type HandlerV1 struct {
	logger log.Logger
	store  StoreV1
}

func NewHandlerV1(
	logger log.Logger,
	store StoreV1,
) *HandlerV1 {
	return &HandlerV1{
		logger: logger.Scoped("codyaccess.v1"),
		store:  store,
	}
}

var _ codyaccessv1connect.CodyAccessServiceHandler = (*HandlerV1)(nil)

func (s *HandlerV1) GetCodyGatewayAccess(ctx context.Context, req *connect.Request[codyaccessv1.GetCodyGatewayAccessRequest]) (*connect.Response[codyaccessv1.GetCodyGatewayAccessResponse], error) {
	logger := trace.Logger(ctx, s.logger).
		With(log.String("queryType", fmt.Sprintf("%T", req.Msg.GetQuery())))

	// 🚨 SECURITY: Require approrpiate M2M scope.
	requiredScope := samsm2m.EnterprisePortalScope(
		scopes.PermissionEnterprisePortalCodyAccess, scopes.ActionRead)
	clientAttrs, err := samsm2m.RequireScope(ctx, logger, s.store, requiredScope, req)
	if err != nil {
		return nil, err
	}
	logger = logger.With(clientAttrs...)

	var attr *codyaccess.CodyGatewayAccessWithSubscriptionDetails
	switch query := req.Msg.GetQuery().(type) {
	case *codyaccessv1.GetCodyGatewayAccessRequest_SubscriptionId:
		if len(query.SubscriptionId) == 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid query: subscription ID"))
		}
		attr, err = s.store.GetCodyGatewayAccessBySubscription(ctx, query.SubscriptionId)

	case *codyaccessv1.GetCodyGatewayAccessRequest_AccessToken:
		if len(query.AccessToken) == 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid query: access token"))
		}
		attr, err = s.store.GetCodyGatewayAccessByAccessToken(ctx, query.AccessToken)

	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid query"))
	}
	if err != nil {
		if errors.Is(err, codyaccess.ErrSubscriptionNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connectutil.InternalError(ctx, logger, err,
			"failed to get Cody Gateway access attributes")
	}

	access := convertAccessAttrsToProto(attr)
	logger.Scoped("audit").Info("GetCodyGatewayAccess",
		log.String("accessedSubscription", access.GetSubscriptionId()))
	return connect.NewResponse(&codyaccessv1.GetCodyGatewayAccessResponse{
		Access: access,
	}), nil
}

func (s *HandlerV1) ListCodyGatewayAccesses(ctx context.Context, req *connect.Request[codyaccessv1.ListCodyGatewayAccessesRequest]) (*connect.Response[codyaccessv1.ListCodyGatewayAccessesResponse], error) {
	logger := trace.Logger(ctx, s.logger)

	// 🚨 SECURITY: Require approrpiate M2M scope.
	requiredScope := samsm2m.EnterprisePortalScope(
		scopes.PermissionEnterprisePortalCodyAccess, scopes.ActionRead)
	clientAttrs, err := samsm2m.RequireScope(ctx, logger, s.store, requiredScope, req)
	if err != nil {
		return nil, err
	}
	logger = logger.With(clientAttrs...)

	// Pagination is unimplemented: https://linear.app/sourcegraph/issue/CORE-134
	if req.Msg.PageSize != 0 {
		return nil, connect.NewError(connect.CodeUnimplemented, errors.New("pagination not implemented"))
	}
	if req.Msg.PageToken != "" {
		return nil, connect.NewError(connect.CodeUnimplemented, errors.New("pagination not implemented"))
	}

	attrs, err := s.store.ListCodyGatewayAccesses(ctx)
	if err != nil {
		if err == dotcomdb.ErrCodyGatewayAccessNotFound {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connectutil.InternalError(ctx, logger, err,
			"failed to list Cody Gateway access attributes")
	}
	resp := codyaccessv1.ListCodyGatewayAccessesResponse{
		// Never a next page, pagination is not implemented yet:
		// https://linear.app/sourcegraph/issue/CORE-134
		NextPageToken: "",
		Accesses:      make([]*codyaccessv1.CodyGatewayAccess, len(attrs)),
	}
	accessedSubscriptions := make([]string, len(attrs))
	for i, attr := range attrs {
		resp.Accesses[i] = convertAccessAttrsToProto(attr)
		accessedSubscriptions[i] = resp.Accesses[i].GetSubscriptionId()
	}
	logger.Scoped("audit").Info("ListCodyGatewayAccesses",
		log.Strings("accessedSubscriptions", accessedSubscriptions))
	return connect.NewResponse(&resp), nil
}

func (s *HandlerV1) UpdateCodyGatewayAccess(ctx context.Context, req *connect.Request[codyaccessv1.UpdateCodyGatewayAccessRequest]) (*connect.Response[codyaccessv1.UpdateCodyGatewayAccessResponse], error) {
	logger := trace.Logger(ctx, s.logger)

	// 🚨 SECURITY: Require approrpiate M2M scope.
	requiredScope := samsm2m.EnterprisePortalScope(
		scopes.PermissionEnterprisePortalCodyAccess, scopes.ActionWrite)
	clientAttrs, err := samsm2m.RequireScope(ctx, logger, s.store, requiredScope, req)
	if err != nil {
		return nil, err
	}
	logger = logger.With(clientAttrs...)

	subscriptionID := req.Msg.GetAccess().SubscriptionId
	if subscriptionID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("subscription ID is required"))
	}

	var opts codyaccess.UpsertCodyGatewayAccessOptions

	update := req.Msg.GetAccess()
	fieldPaths := req.Msg.GetUpdateMask().GetPaths()
	// Empty field paths means update all non-empty fields.
	if len(fieldPaths) == 0 {
		if update.Enabled {
			opts.Enabled = pointers.Ptr(update.Enabled)
		}
		if update.GetChatCompletionsRateLimit().GetLimit() > 0 {
			opts.ChatCompletionsRateLimit = database.NewNullInt64(
				update.GetChatCompletionsRateLimit().Limit)
		}
		if update.GetChatCompletionsRateLimit().GetIntervalDuration().GetSeconds() > 0 {
			opts.ChatCompletionsRateLimitIntervalSeconds = database.NewNullInt32(
				update.GetChatCompletionsRateLimit().GetIntervalDuration().Seconds)
		}
		if update.GetCodeCompletionsRateLimit().GetLimit() > 0 {
			opts.CodeCompletionsRateLimit = database.NewNullInt64(
				update.GetCodeCompletionsRateLimit().Limit)
		}
		if update.GetCodeCompletionsRateLimit().GetIntervalDuration().GetSeconds() > 0 {
			opts.CodeCompletionsRateLimitIntervalSeconds = database.NewNullInt32(
				update.GetCodeCompletionsRateLimit().GetIntervalDuration().Seconds)
		}
		if update.GetEmbeddingsRateLimit().GetLimit() > 0 {
			opts.EmbeddingsRateLimit = database.NewNullInt64(
				update.GetEmbeddingsRateLimit().Limit)
		}
		if update.GetEmbeddingsRateLimit().GetIntervalDuration().GetSeconds() > 0 {
			opts.EmbeddingsRateLimitIntervalSeconds = database.NewNullInt32(
				update.GetEmbeddingsRateLimit().GetIntervalDuration().Seconds)
		}
	} else {
		for _, p := range fieldPaths {
			var valid bool
			if p == "*" {
				valid = true
				opts.ForceUpdate = true
			}
			if p == "enabled" || p == "*" {
				valid = true
				opts.Enabled = pointers.Ptr(update.GetEnabled())
			}
			if p == "chat_completions_rate_limit.limit" || p == "*" {
				valid = true
				if update.ChatCompletionsRateLimit == nil {
					opts.ChatCompletionsRateLimit = &sql.NullInt64{}
				} else {
					opts.ChatCompletionsRateLimit = database.NewNullInt64(
						update.GetChatCompletionsRateLimit().GetLimit())
				}
			}
			if p == "chat_completions_rate_limit.interval_duration" || p == "*" {
				valid = true
				if update.ChatCompletionsRateLimit == nil {
					opts.ChatCompletionsRateLimitIntervalSeconds = &sql.NullInt32{}
				} else {
					opts.ChatCompletionsRateLimitIntervalSeconds = database.NewNullInt32(
						update.GetChatCompletionsRateLimit().GetIntervalDuration().GetSeconds())
				}
			}
			if p == "code_completions_rate_limit.limit" || p == "*" {
				valid = true
				if update.CodeCompletionsRateLimit == nil {
					opts.CodeCompletionsRateLimit = &sql.NullInt64{}
				} else {
					opts.CodeCompletionsRateLimit = database.NewNullInt64(
						update.GetCodeCompletionsRateLimit().GetLimit())
				}
			}
			if p == "code_completions_rate_limit.interval_duration" || p == "*" {
				valid = true
				if update.CodeCompletionsRateLimit == nil {
					opts.CodeCompletionsRateLimitIntervalSeconds = &sql.NullInt32{}
				} else {
					opts.CodeCompletionsRateLimitIntervalSeconds = database.NewNullInt32(
						update.GetCodeCompletionsRateLimit().GetIntervalDuration().GetSeconds())
				}
			}
			if p == "embeddings_rate_limit.limit" || p == "*" {
				valid = true
				if update.EmbeddingsRateLimit == nil {
					opts.EmbeddingsRateLimit = &sql.NullInt64{}
				} else {
					opts.EmbeddingsRateLimit = database.NewNullInt64(
						update.GetEmbeddingsRateLimit().GetLimit())
				}
			}
			if p == "embeddings_rate_limit.interval_duration" || p == "*" {
				valid = true
				if update.EmbeddingsRateLimit == nil {
					opts.EmbeddingsRateLimitIntervalSeconds = &sql.NullInt32{}
				} else {
					opts.EmbeddingsRateLimitIntervalSeconds = database.NewNullInt32(
						update.GetEmbeddingsRateLimit().GetIntervalDuration().GetSeconds())
				}
			}
			if !valid {
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.Newf("invalid field path %q", p))
			}
		}
	}

	updated, err := s.store.UpsertCodyGatewayAccess(ctx, subscriptionID, opts)
	if err != nil {
		if errors.Is(err, codyaccess.ErrSubscriptionNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connectutil.InternalError(ctx, logger, err, "failed to update Cody Gateway access")
	}

	logger.Scoped("audit").Info("GetCodyGatewayAccess",
		log.String("updatedSubscription", subscriptionID))

	return connect.NewResponse(&codyaccessv1.UpdateCodyGatewayAccessResponse{
		Access: convertAccessAttrsToProto(updated),
	}), nil
}

func (s *HandlerV1) GetCodyGatewayUsage(ctx context.Context, req *connect.Request[codyaccessv1.GetCodyGatewayUsageRequest]) (*connect.Response[codyaccessv1.GetCodyGatewayUsageResponse], error) {
	logger := trace.Logger(ctx, s.logger)

	// 🚨 SECURITY: Require appropriate M2M scope.
	requiredScope := samsm2m.EnterprisePortalScope(
		scopes.PermissionEnterprisePortalCodyAccess, scopes.ActionRead)
	clientAttrs, err := samsm2m.RequireScope(ctx, logger, s.store, requiredScope, req)
	if err != nil {
		return nil, err
	}
	logger = logger.With(clientAttrs...)

	switch query := req.Msg.GetQuery().(type) {
	case *codyaccessv1.GetCodyGatewayUsageRequest_SubscriptionId:
		internalSubscriptionID := strings.TrimPrefix(query.SubscriptionId,
			subscriptionsv1.EnterpriseSubscriptionIDPrefix)
		if internalSubscriptionID == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid query: subscription ID"))
		}

		usage, err := s.store.GetCodyGatewayUsage(ctx, internalSubscriptionID)
		if err != nil {
			if errors.Is(err, errStoreUnimplemented) {
				return nil, connect.NewError(connect.CodeUnimplemented, err)
			} else if errors.IsContextCanceled(err) {
				return nil, connect.NewError(connect.CodeAborted, err)
			}
			return nil, connectutil.InternalError(ctx, logger, err, "failed to get Cody Gateway usage")
		}
		return connect.NewResponse(&codyaccessv1.GetCodyGatewayUsageResponse{
			Usage: usage,
		}), nil

	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Newf("unknown query type %T", query))
	}
}
