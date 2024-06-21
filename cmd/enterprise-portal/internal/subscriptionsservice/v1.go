package subscriptionsservice

import (
	"context"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/sourcegraph/log"
	"golang.org/x/exp/maps"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"

	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
	subscriptionsv1connect "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1/v1connect"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/iam"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/connectutil"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/samsm2m"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

const Name = subscriptionsv1connect.SubscriptionsServiceName

func RegisterV1(
	logger log.Logger,
	mux *http.ServeMux,
	store StoreV1,
	opts ...connect.HandlerOption,
) {
	mux.Handle(
		subscriptionsv1connect.NewSubscriptionsServiceHandler(
			&handlerV1{
				logger: logger.Scoped("subscriptions.v1"),
				store:  store,
			},
			opts...,
		),
	)
}

type handlerV1 struct {
	subscriptionsv1connect.UnimplementedSubscriptionsServiceHandler

	logger log.Logger
	store  StoreV1
}

var _ subscriptionsv1connect.SubscriptionsServiceHandler = (*handlerV1)(nil)

func (s *handlerV1) ListEnterpriseSubscriptions(ctx context.Context, req *connect.Request[subscriptionsv1.ListEnterpriseSubscriptionsRequest]) (*connect.Response[subscriptionsv1.ListEnterpriseSubscriptionsResponse], error) {
	logger := trace.Logger(ctx, s.logger)

	// 🚨 SECURITY: Require appropriate M2M scope.
	requiredScope := samsm2m.EnterprisePortalScope("subscription", scopes.ActionRead)
	clientAttrs, err := samsm2m.RequireScope(ctx, logger, s.store, requiredScope, req)
	if err != nil {
		return nil, err
	}
	logger = logger.With(clientAttrs...)

	// Pagination is unimplemented (https://linear.app/sourcegraph/issue/CORE-134),
	// but we do want to allow pageSize to act as a 'limit' parameter for querying
	// product subscriptions.
	if req.Msg.PageToken != "" {
		return nil, connect.NewError(connect.CodeUnimplemented, errors.New("pagination not implemented"))
	}

	// Validate and process filters.
	filters := req.Msg.GetFilters()
	isArchived := false
	subscriptionIDs := make(collections.Set[string], len(filters))
	var iamListObjectOptions *iam.ListObjectsOptions
	for _, filter := range filters {
		switch f := filter.GetFilter().(type) {
		case *subscriptionsv1.ListEnterpriseSubscriptionsFilter_SubscriptionId:
			if f.SubscriptionId == "" {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.New(`invalid filter: "subscription_id" provided but is empty`),
				)
			}
			subscriptionIDs.Add(
				strings.TrimPrefix(f.SubscriptionId, subscriptionsv1.EnterpriseSubscriptionIDPrefix))
		case *subscriptionsv1.ListEnterpriseSubscriptionsFilter_IsArchived:
			isArchived = f.IsArchived
		case *subscriptionsv1.ListEnterpriseSubscriptionsFilter_Permission:
			if f.Permission == nil {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.New(`invalid filter: "permission" provided but is empty`),
				)
			} else if f.Permission.Type == subscriptionsv1.PermissionType_PERMISSION_TYPE_UNSPECIFIED {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.New(`invalid filter: "permission" provided but "type" is unspecified`),
				)
			} else if f.Permission.Relation == subscriptionsv1.PermissionRelation_PERMISSION_RELATION_UNSPECIFIED {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.New(`invalid filter: "permission" provided but "relation" is unspecified`),
				)
			} else if f.Permission.SamsAccountId == "" {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.New(`invalid filter: "permission" provided but "sams_account_id" is empty`),
				)
			}

			if iamListObjectOptions != nil {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.New(`invalid filter: "permission" can only be provided once`),
				)
			}
			iamListObjectOptions = &iam.ListObjectsOptions{
				Type:     convertProtoToIAMTupleObjectType(f.Permission.Type),
				Relation: convertProtoToIAMTupleRelation(f.Permission.Relation),
				Subject:  iam.ToTupleSubjectUser(f.Permission.SamsAccountId),
			}
		}
	}

	// When requested, evaluate the list of subscriptions that match the
	// filtered permission.
	if iamListObjectOptions != nil {
		// Object IDs are in the form of `subscription_cody_analytics:<subscriptionID>`.
		objectIDs, err := s.store.IAMListObjects(ctx, *iamListObjectOptions)
		if err != nil {
			return nil, connectutil.InternalError(ctx, logger, err, "list subscriptions from IAM")
		}
		allowedSubscriptionIDs := make(collections.Set[string], len(objectIDs))
		for _, objectID := range objectIDs {
			allowedSubscriptionIDs.Add(strings.TrimPrefix(objectID, "subscription_cody_analytics:"))
		}

		if !subscriptionIDs.IsEmpty() {
			// If subscription IDs were provided, we only want to return the
			// subscriptions that are part of the provided IDs.
			subscriptionIDs = collections.Intersection(subscriptionIDs, allowedSubscriptionIDs)
		} else {
			// Otherwise, only return the allowed subscriptions.
			subscriptionIDs = allowedSubscriptionIDs
		}

		// 🚨 SECURITY: If permissions are used as filter, but we found no results, we
		// should directly return an empty response to not mistaken as list all.
		if len(subscriptionIDs) == 0 {
			return connect.NewResponse(&subscriptionsv1.ListEnterpriseSubscriptionsResponse{}), nil
		}
	}

	subscriptions, err := s.store.ListEnterpriseSubscriptions(
		ctx,
		database.ListEnterpriseSubscriptionsOptions{
			IDs:        subscriptionIDs.Values(),
			IsArchived: isArchived,
			PageSize:   int(req.Msg.GetPageSize()),
		},
	)
	if err != nil {
		return nil, connectutil.InternalError(ctx, logger, err, "list subscriptions")
	}

	// List from dotcom DB and merge attributes.
	dotcomSubscriptions, err := s.store.ListDotcomEnterpriseSubscriptions(ctx, dotcomdb.ListEnterpriseSubscriptionsOptions{
		SubscriptionIDs: subscriptionIDs.Values(),
		IsArchived:      isArchived,
	})
	if err != nil {
		return nil, connectutil.InternalError(ctx, logger, err, "list subscriptions from dotcom DB")
	}
	dotcomSubscriptionsSet := make(map[string]*dotcomdb.SubscriptionAttributes, len(dotcomSubscriptions))
	for _, s := range dotcomSubscriptions {
		dotcomSubscriptionsSet[s.ID] = s
	}

	// Add the "real" subscriptions we already track to the results
	respSubscriptions := make([]*subscriptionsv1.EnterpriseSubscription, 0, len(subscriptions))
	for _, s := range subscriptions {
		respSubscriptions = append(
			respSubscriptions,
			convertSubscriptionToProto(s, dotcomSubscriptionsSet[s.ID]),
		)
		delete(dotcomSubscriptionsSet, s.ID)
	}

	// Add any remaining dotcom subscriptions to the results set
	for _, s := range dotcomSubscriptionsSet {
		respSubscriptions = append(
			respSubscriptions,
			convertSubscriptionToProto(&database.Subscription{
				ID: subscriptionsv1.EnterpriseSubscriptionIDPrefix + s.ID,
			}, s),
		)
	}

	accessedSubscriptions := map[string]struct{}{}
	for _, s := range respSubscriptions {
		accessedSubscriptions[s.GetId()] = struct{}{}
	}
	logger.Scoped("audit").Info("ListEnterpriseSubscriptions",
		log.Strings("accessedSubscriptions", maps.Keys(accessedSubscriptions)),
	)

	return connect.NewResponse(
		&subscriptionsv1.ListEnterpriseSubscriptionsResponse{
			// Never a next page, pagination is not implemented yet:
			// https://linear.app/sourcegraph/issue/CORE-134
			NextPageToken: "",
			Subscriptions: respSubscriptions,
		},
	), nil
}

func (s *handlerV1) ListEnterpriseSubscriptionLicenses(ctx context.Context, req *connect.Request[subscriptionsv1.ListEnterpriseSubscriptionLicensesRequest]) (*connect.Response[subscriptionsv1.ListEnterpriseSubscriptionLicensesResponse], error) {
	logger := trace.Logger(ctx, s.logger)

	// 🚨 SECURITY: Require appropriate M2M scope.
	requiredScope := samsm2m.EnterprisePortalScope("subscription", scopes.ActionRead)
	clientAttrs, err := samsm2m.RequireScope(ctx, logger, s.store, requiredScope, req)
	if err != nil {
		return nil, err
	}
	logger = logger.With(clientAttrs...)

	// Pagination is unimplemented: https://linear.app/sourcegraph/issue/CORE-134
	// BUT, we allow pageSize to act as a 'limit' parameter for querying for
	// 'active license'.
	if req.Msg.PageToken != "" {
		return nil, connect.NewError(connect.CodeUnimplemented, errors.New("pagination not implemented"))
	}

	// Validate filters
	filters := req.Msg.GetFilters()
	for _, filter := range filters {
		// TODO: Implement additional filtering as needed
		switch f := filter.GetFilter().(type) {
		case *subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_Type:
			return nil, connect.NewError(connect.CodeUnimplemented,
				errors.New("filtering by type is not implemented"))
		case *subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_LicenseKeySubstring:
			return nil, connect.NewError(connect.CodeUnimplemented,
				errors.New("filtering by license key substring is not implemented"))
		case *subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_SubscriptionId:
			if f.SubscriptionId == "" {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.New(`invalid filter: "subscription_id"" provided but is empty`),
				)
			}
		}
	}

	licenses, err := s.store.ListDotcomEnterpriseSubscriptionLicenses(ctx, filters,
		// Provide page size to allow "active license" functionality, by only
		// retrieving the most recently created result.
		int(req.Msg.GetPageSize()))
	if err != nil {
		if errors.Is(err, dotcomdb.ErrCodyGatewayAccessNotFound) {
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

	accessedSubscriptions := map[string]struct{}{}
	accessedLicenses := make([]string, len(licenses))
	for i, l := range licenses {
		resp.Licenses[i] = convertLicenseAttrsToProto(l)
		accessedSubscriptions[resp.Licenses[i].GetSubscriptionId()] = struct{}{}
		accessedLicenses[i] = resp.Licenses[i].GetId()
	}
	logger.Scoped("audit").Info("ListEnterpriseSubscriptionLicenses",
		log.Strings("accessedSubscriptions", maps.Keys(accessedSubscriptions)),
		log.Strings("accessedLicenses", accessedLicenses))
	return connect.NewResponse(&resp), nil
}

func (s *handlerV1) UpdateEnterpriseSubscription(ctx context.Context, req *connect.Request[subscriptionsv1.UpdateEnterpriseSubscriptionRequest]) (*connect.Response[subscriptionsv1.UpdateEnterpriseSubscriptionResponse], error) {
	logger := trace.Logger(ctx, s.logger)

	// 🚨 SECURITY: Require appropriate M2M scope.
	requiredScope := samsm2m.EnterprisePortalScope("subscription", scopes.ActionWrite)
	clientAttrs, err := samsm2m.RequireScope(ctx, logger, s.store, requiredScope, req)
	if err != nil {
		return nil, err
	}
	logger = logger.With(clientAttrs...)

	subscriptionID := strings.TrimPrefix(req.Msg.GetSubscription().GetId(), subscriptionsv1.EnterpriseSubscriptionIDPrefix)
	if subscriptionID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("subscription.id is required"))
	}

	// Double check with the dotcom DB that the subscription ID is valid.
	subscriptionAttrs, err := s.store.ListDotcomEnterpriseSubscriptions(ctx, dotcomdb.ListEnterpriseSubscriptionsOptions{
		SubscriptionIDs: []string{subscriptionID},
	})
	if err != nil {
		return nil, connectutil.InternalError(ctx, logger, err, "get dotcom enterprise subscription")
	} else if len(subscriptionAttrs) != 1 {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("subscription not found"))
	}

	var opts database.UpsertSubscriptionOptions

	fieldPaths := req.Msg.GetUpdateMask().GetPaths()
	// Empty field paths means update all non-empty fields.
	if len(fieldPaths) == 0 {
		if v := req.Msg.GetSubscription().GetInstanceDomain(); v != "" {
			opts.InstanceDomain = v
		}
	} else {
		for _, p := range fieldPaths {
			switch p {
			case "instance_domain":
				opts.InstanceDomain = req.Msg.GetSubscription().GetInstanceDomain()
			case "*":
				opts.ForceUpdate = true
				opts.InstanceDomain = req.Msg.GetSubscription().GetInstanceDomain()
			}
		}
	}

	// Validate and normalize the domain
	if opts.InstanceDomain != "" {
		opts.InstanceDomain, err = subscriptionsv1.NormalizeInstanceDomain(opts.InstanceDomain)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid instance domain"))
		}
	}

	subscription, err := s.store.UpsertEnterpriseSubscription(ctx, subscriptionID, opts)
	if err != nil {
		return nil, connectutil.InternalError(ctx, logger, err, "upsert subscription")
	}

	respSubscription := convertSubscriptionToProto(subscription, subscriptionAttrs[0])
	logger.Scoped("audit").Info("UpdateEnterpriseSubscription",
		log.String("updatedSubscription", respSubscription.GetId()),
	)

	return connect.NewResponse(
		&subscriptionsv1.UpdateEnterpriseSubscriptionResponse{
			Subscription: respSubscription,
		},
	), nil
}

func (s *handlerV1) UpdateEnterpriseSubscriptionMembership(ctx context.Context, req *connect.Request[subscriptionsv1.UpdateEnterpriseSubscriptionMembershipRequest]) (*connect.Response[subscriptionsv1.UpdateEnterpriseSubscriptionMembershipResponse], error) {
	logger := trace.Logger(ctx, s.logger)

	// 🚨 SECURITY: Require appropriate M2M scope.
	requiredScope := samsm2m.EnterprisePortalScope("permission.subscription", scopes.ActionWrite)
	clientAttrs, err := samsm2m.RequireScope(ctx, logger, s.store, requiredScope, req)
	if err != nil {
		return nil, err
	}
	logger = logger.With(clientAttrs...)

	samsAccountID := req.Msg.GetMembership().GetMemberSamsAccountId()
	if samsAccountID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("membership.member_sams_account_id is required"))
	}

	_, err = s.store.GetSAMSUserByID(ctx, samsAccountID)
	if err != nil {
		if errors.Is(err, sams.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("SAMS user not found"))
		}
		return nil, connectutil.InternalError(ctx, logger, err, "get SAMS user by ID")
	}

	subscriptionID := strings.TrimPrefix(req.Msg.GetMembership().GetSubscriptionId(), subscriptionsv1.EnterpriseSubscriptionIDPrefix)
	instanceDomain := req.Msg.GetMembership().GetInstanceDomain()
	if subscriptionID == "" && instanceDomain == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("either membership.subscription_id or membership.instance_domain must be specified"),
		)
	}

	if subscriptionID != "" {
		// Double check with the dotcom DB that the subscription ID is valid.
		subscriptionAttrs, err := s.store.ListDotcomEnterpriseSubscriptions(ctx, dotcomdb.ListEnterpriseSubscriptionsOptions{
			SubscriptionIDs: []string{subscriptionID},
		})
		if err != nil {
			return nil, connectutil.InternalError(ctx, logger, err, "get dotcom enterprise subscription")
		} else if len(subscriptionAttrs) != 1 {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("subscription not found"))
		}
	} else if instanceDomain != "" {
		// Validate and normalize the domain
		instanceDomain, err = subscriptionsv1.NormalizeInstanceDomain(instanceDomain)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid instance domain"))
		}
		subscriptions, err := s.store.ListEnterpriseSubscriptions(
			ctx,
			database.ListEnterpriseSubscriptionsOptions{
				InstanceDomains: []string{instanceDomain},
				PageSize:        1,
			},
		)
		if err != nil {
			return nil, connectutil.InternalError(ctx, logger, err, "list subscriptions")
		} else if len(subscriptions) != 1 {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("subscription not found"))
		}
		subscriptionID = subscriptions[0].ID
	}

	roles := req.Msg.GetMembership().GetMemberRoles()
	writes := make([]iam.TupleKey, 0, len(roles))
	for _, role := range roles {
		switch role {
		case subscriptionsv1.Role_ROLE_SUBSCRIPTION_CODY_ANALYTICS_CUSTOMER_ADMIN:
			// Subscription cody analytics customer admin can:
			//	- View cody analytics of the subscription

			// Make sure the customer_admin role is created for the subscription.
			tk := iam.TupleKey{
				Object:        iam.ToTupleObject(iam.TupleTypeSubscriptionCodyAnalytics, subscriptionID),
				TupleRelation: iam.TupleRelationView,
				Subject:       iam.ToTupleSubjectCustomerAdmin(subscriptionID, iam.TupleRelationMember),
			}
			allowed, err := s.store.IAMCheck(ctx, iam.CheckOptions{TupleKey: tk})
			if err != nil {
				return nil, connectutil.InternalError(ctx, logger, err, "check relation tuple in IAM")
			}
			if !allowed {
				writes = append(writes, tk)
			}

			// Add the user as a member of the customer_admin role of the subscription.
			tk = iam.TupleKey{
				Object:        iam.ToTupleObject(iam.TupleTypeCustomerAdmin, subscriptionID),
				TupleRelation: iam.TupleRelationMember,
				Subject:       iam.ToTupleSubjectUser(samsAccountID),
			}
			allowed, err = s.store.IAMCheck(ctx, iam.CheckOptions{TupleKey: tk})
			if err != nil {
				return nil, connectutil.InternalError(ctx, logger, err, "check relation tuple in IAM")
			}
			if !allowed {
				writes = append(writes, tk)
			}
		default:
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.Errorf("invalid role: %s", role))
		}
	}

	err = s.store.IAMWrite(
		ctx,
		iam.WriteOptions{
			Writes: writes,
		},
	)
	if err != nil {
		return nil, connectutil.InternalError(ctx, logger, err, "write relation tuples to IAM")
	}
	return connect.NewResponse(&subscriptionsv1.UpdateEnterpriseSubscriptionMembershipResponse{}), nil
}
