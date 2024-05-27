package subscriptionsservice

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/connectutil"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/samsm2m"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
	subscriptionsv1connect "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1/v1connect"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const Name = subscriptionsv1connect.SubscriptionsServiceName

type DotComDB interface {
	ListEnterpriseSubscriptionLicenses(context.Context, []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter, int) ([]*dotcomdb.LicenseAttributes, error)
}

func RegisterV1(logger log.Logger, mux *http.ServeMux, samsClient samsm2m.TokenIntrospector, dotcom DotComDB) {
	mux.Handle(subscriptionsv1connect.NewSubscriptionsServiceHandler(&handlerV1{
		logger:     logger.Scoped("subscriptions.v1"),
		samsClient: samsClient,
		dotcom:     dotcom,
	}))
}

type handlerV1 struct {
	subscriptionsv1connect.UnimplementedSubscriptionsServiceHandler

	logger     log.Logger
	samsClient samsm2m.TokenIntrospector
	dotcom     DotComDB
}

var _ subscriptionsv1connect.SubscriptionsServiceHandler = (*handlerV1)(nil)

func (s *handlerV1) ListEnterpriseSubscriptionLicenses(ctx context.Context, req *connect.Request[subscriptionsv1.ListEnterpriseSubscriptionLicensesRequest]) (*connect.Response[subscriptionsv1.ListEnterpriseSubscriptionLicensesResponse], error) {
	logger := trace.Logger(ctx, s.logger)

	// 🚨 SECURITY: Require approrpiate M2M scope.
	requiredScope := samsm2m.EnterprisePortalScope("subscription", scopes.ActionRead)
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

	// Validate filters
	filters := req.Msg.GetFilters()
	if len(filters) == 0 {
		// TODO: We may want to allow filter-less usage in the future
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("at least one filter is required"))
	}
	for _, f := range filters {
		// TODO: Implement additional filtering as needed
		switch f.Filter.(type) {
		case *subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_Type:
			return nil, connect.NewError(connect.CodeUnimplemented,
				errors.New("filtering by type is not implemented"))
		case *subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_LicenseKeySubstring:
			return nil, connect.NewError(connect.CodeUnimplemented,
				errors.New("filtering by license key substring is not implemented"))
		}
	}

	licenses, err := s.dotcom.ListEnterpriseSubscriptionLicenses(ctx, filters,
		// Provide page size to allow "active license" functionality, by only
		// retrieving the most recently created result.
		int(req.Msg.GetPageSize()))
	if err != nil {
		if err == dotcomdb.ErrCodyGatewayAccessNotFound {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connectutil.InternalError(ctx, logger, err,
			"failed to get Enterprise Subscription licenses")
	}

	resp := subscriptionsv1.ListEnterpriseSubscriptionLicensesResponse{
		// Never a next page, pagination is not implemented yet:
		// https://linear.app/sourcegraph/issue/CORE-134
		NextPageToken: "",
		Licenses:      make([]*subscriptionsv1.EnterpriseSubscriptionLicense, len(licenses)),
	}
	for i, l := range licenses {
		resp.Licenses[i] = convertLicenseAttrsToProto(l)
	}
	return connect.NewResponse(&resp), nil
}
