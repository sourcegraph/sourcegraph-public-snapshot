package codyaccessservice

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/connectutil"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/samsm2m"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	codyaccessv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/codyaccess/v1"
	codyaccessv1connect "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/codyaccess/v1/v1connect"
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
	samsClient samsm2m.TokenIntrospector,
	dotcom DotComDB,
	opts ...connect.HandlerOption,
) {
	mux.Handle(
		codyaccessv1connect.NewCodyAccessServiceHandler(
			&handlerV1{
				logger:     logger.Scoped("codyaccess.v1"),
				samsClient: samsClient,
				dotcom:     dotcom,
			},
			opts...,
		),
	)
}

type handlerV1 struct {
	codyaccessv1connect.UnimplementedCodyAccessServiceHandler
	logger log.Logger

	samsClient samsm2m.TokenIntrospector
	dotcom     DotComDB
}

var _ codyaccessv1connect.CodyAccessServiceHandler = (*handlerV1)(nil)

func (s *handlerV1) GetCodyGatewayAccess(ctx context.Context, req *connect.Request[codyaccessv1.GetCodyGatewayAccessRequest]) (*connect.Response[codyaccessv1.GetCodyGatewayAccessResponse], error) {
	logger := trace.Logger(ctx, s.logger).
		With(log.String("queryType", fmt.Sprintf("%T", req.Msg.GetQuery())))

	// 🚨 SECURITY: Require approrpiate M2M scope.
	requiredScope := samsm2m.EnterprisePortalScope("codyaccess", scopes.ActionRead)
	if err := samsm2m.RequireScope(ctx, logger, s.samsClient, requiredScope, req); err != nil {
		return nil, err
	}

	var attr *dotcomdb.CodyGatewayAccessAttributes
	var err error
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
	return connect.NewResponse(&codyaccessv1.GetCodyGatewayAccessResponse{
		Access: convertAccessAttrsToProto(attr),
	}), nil
}

func (s *handlerV1) ListCodyGatewayAccesses(ctx context.Context, req *connect.Request[codyaccessv1.ListCodyGatewayAccessesRequest]) (*connect.Response[codyaccessv1.ListCodyGatewayAccessesResponse], error) {
	logger := trace.Logger(ctx, s.logger)

	// 🚨 SECURITY: Require approrpiate M2M scope.
	requiredScope := samsm2m.EnterprisePortalScope("codyaccess", scopes.ActionRead)
	if err := samsm2m.RequireScope(ctx, logger, s.samsClient, requiredScope, req); err != nil {
		return nil, err
	}

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
	for i, attr := range attrs {
		resp.Accesses[i] = convertAccessAttrsToProto(attr)
	}
	return connect.NewResponse(&resp), nil
}
