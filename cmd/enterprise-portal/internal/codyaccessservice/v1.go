package codyaccessservice

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/connectutil"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	codyaccessv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/codyaccess/v1"
	codyaccessv1connect "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/codyaccess/v1/v1connect"
)

type DotComDB interface {
	GetCodyGatewayAccessAttributesBySubscription(ctx context.Context, subscriptionID string) (*dotcomdb.CodyGatewayAccessAttributes, error)
	GetCodyGatewayAccessAttributesByAccessToken(ctx context.Context, subscriptionID string) (*dotcomdb.CodyGatewayAccessAttributes, error)
	GetAllCodyGatewayAccessAttributes(ctx context.Context) ([]*dotcomdb.CodyGatewayAccessAttributes, error)
}

func RegisterV1(logger log.Logger, mux *http.ServeMux, dotcom DotComDB) {
	mux.Handle(codyaccessv1connect.NewCodyAccessServiceHandler(newHandlerV1(logger, dotcom)))
}

type handlerV1 struct {
	codyaccessv1connect.UnimplementedCodyAccessServiceHandler

	logger log.Logger
	dotcom DotComDB
}

var _ codyaccessv1connect.CodyAccessServiceHandler = (*handlerV1)(nil)

// newHandlerV1 implements enterpriseportal/codyaccess/v1/v1connect.
func newHandlerV1(logger log.Logger, dotcom DotComDB) *handlerV1 {
	return &handlerV1{
		logger: logger.Scoped("codyaccess.v1"),
		dotcom: dotcom,
	}
}

func (s *handlerV1) GetCodyGatewayAccess(ctx context.Context, req *connect.Request[codyaccessv1.GetCodyGatewayAccessRequest]) (*connect.Response[codyaccessv1.GetCodyGatewayAccessResponse], error) {
	logger := trace.Logger(ctx, s.logger).
		With(log.String("query", fmt.Sprintf("%T", req.Msg.GetQuery())))

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
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid query: subscription ID"))
		}
		attr, err = s.dotcom.GetCodyGatewayAccessAttributesByAccessToken(ctx, query.AccessToken)

	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid query"))
	}
	if err != nil {
		if err == dotcomdb.ErrCodyGatewayAccessNotFound {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connectutil.InternalError(logger, err,
			"failed to get Cody Gateway access attributes")
	}
	return connect.NewResponse(&codyaccessv1.GetCodyGatewayAccessResponse{
		Access: convertAccessAttrsToProto(attr),
	}), nil
}

func (s *handlerV1) ListCodyGatewayAccesses(ctx context.Context, req *connect.Request[codyaccessv1.ListCodyGatewayAccessesRequest]) (*connect.Response[codyaccessv1.ListCodyGatewayAccessesResponse], error) {
	logger := trace.Logger(ctx, s.logger)

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
		return nil, connectutil.InternalError(logger, err,
			"failed to list Cody Gateway access attributes")
	}
	resp := codyaccessv1.ListCodyGatewayAccessesResponse{
		NextPageToken: "", // never a next page, pagination is not implemented yet
		Accesses:      make([]*codyaccessv1.CodyGatewayAccess, len(attrs)),
	}
	for i, attr := range attrs {
		resp.Accesses[i] = convertAccessAttrsToProto(attr)
	}
	return connect.NewResponse(&resp), nil
}
