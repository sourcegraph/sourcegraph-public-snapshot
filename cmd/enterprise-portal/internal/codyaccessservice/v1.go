package codyaccessservice

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/connectutil"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/samsm2m"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	codyaccessv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/codyaccess/v1"
	codyaccessv1connect "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/codyaccess/v1/v1connect"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const Name = codyaccessv1connect.CodyAccessServiceName

type DotComDB interface {
	GetCodyGatewayAccessAttributesBySubscription(ctx context.Context, subscriptionID string) (*dotcomdb.CodyGatewayAccessAttributes, error)
	GetCodyGatewayAccessAttributesByAccessToken(ctx context.Context, subscriptionID string) (*dotcomdb.CodyGatewayAccessAttributes, error)
	GetAllCodyGatewayAccessAttributes(ctx context.Context) ([]*dotcomdb.CodyGatewayAccessAttributes, error)
}

func RegisterV1(
	logger log.Logger,
	mux *http.ServeMux,
	store StoreV1,
	dotcom DotComDB,
	opts ...connect.HandlerOption,
) {
	mux.Handle(
		codyaccessv1connect.NewCodyAccessServiceHandler(
			&handlerV1{
				logger: logger.Scoped("codyaccess.v1"),
				store:  store,
				dotcom: dotcom,
			},
			opts...,
		),
	)
}

type handlerV1 struct {
	codyaccessv1connect.UnimplementedCodyAccessServiceHandler

	logger log.Logger
	store  StoreV1
	dotcom DotComDB
}

var _ codyaccessv1connect.CodyAccessServiceHandler = (*handlerV1)(nil)

func (s *handlerV1) GetCodyGatewayAccess(ctx context.Context, req *connect.Request[codyaccessv1.GetCodyGatewayAccessRequest]) (*connect.Response[codyaccessv1.GetCodyGatewayAccessResponse], error) {
	logger := trace.Logger(ctx, s.logger).
		With(log.String("queryType", fmt.Sprintf("%T", req.Msg.GetQuery())))

	// 🚨 SECURITY: Require approrpiate M2M scope.
	requiredScope := samsm2m.EnterprisePortalScope("codyaccess", scopes.ActionRead)
	clientAttrs, err := samsm2m.RequireScope(ctx, logger, s.store, requiredScope, req)
	if err != nil {
		return nil, err
	}
	logger = logger.With(clientAttrs...)

	var attr *dotcomdb.CodyGatewayAccessAttributes
	switch query := req.Msg.GetQuery().(type) {
	case *codyaccessv1.GetCodyGatewayAccessRequest_SubscriptionId:
		if len(query.SubscriptionId) == 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid query: subscription ID"))
		}
		attr, err = s.dotcom.GetCodyGatewayAccessAttributesBySubscription(ctx, query.SubscriptionId)

	case *codyaccessv1.GetCodyGatewayAccessRequest_AccessToken:
		if len(query.AccessToken) == 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid query: access token"))
		}
		attr, err = s.dotcom.GetCodyGatewayAccessAttributesByAccessToken(ctx, query.AccessToken)

	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid query"))
	}
	if err != nil {
		if errors.Is(err, dotcomdb.ErrCodyGatewayAccessNotFound) {
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

func (s *handlerV1) ListCodyGatewayAccesses(ctx context.Context, req *connect.Request[codyaccessv1.ListCodyGatewayAccessesRequest]) (*connect.Response[codyaccessv1.ListCodyGatewayAccessesResponse], error) {
	logger := trace.Logger(ctx, s.logger)

	// 🚨 SECURITY: Require approrpiate M2M scope.
	requiredScope := samsm2m.EnterprisePortalScope("codyaccess", scopes.ActionRead)
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

	attrs, err := s.dotcom.GetAllCodyGatewayAccessAttributes(ctx)
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

func (s *handlerV1) GetCodyGatewayUsage(ctx context.Context, req *connect.Request[codyaccessv1.GetCodyGatewayUsageRequest]) (*connect.Response[codyaccessv1.GetCodyGatewayUsageResponse], error) {
	logger := trace.Logger(ctx, s.logger)

	// 🚨 SECURITY: Require approrpiate M2M scope.
	requiredScope := samsm2m.EnterprisePortalScope("codyaccess", scopes.ActionRead)
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
