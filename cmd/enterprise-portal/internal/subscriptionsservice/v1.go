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
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
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
	var (
		isArchived                *bool
		subscriptionIDs           = make(collections.Set[string], len(filters))
		displayNameSubstring      string
		salesforceSubscriptionIDs []string
		instanceDomains           []string
	)
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
			isArchived = &f.IsArchived
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
			if err := iamListObjectOptions.Validate(); err != nil {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.Wrap(err, `invalid filter: "permission" provided but invalid`),
				)
			}
		case *subscriptionsv1.ListEnterpriseSubscriptionsFilter_DisplayName:
			if displayNameSubstring != "" {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.Newf(`invalid filter: "display_name" provided more than once`),
				)
			}
			const minLength = 3
			if len(f.DisplayName) < minLength {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.Newf(`invalid filter: "display_name" must be longer than %d characters`, minLength),
				)
			}
			displayNameSubstring = f.DisplayName
		case *subscriptionsv1.ListEnterpriseSubscriptionsFilter_Salesforce:
			if f.Salesforce.SubscriptionId == "" {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.Newf(`invalid filter: "salesforce.subscription_id" is empty`),
				)
			}
			salesforceSubscriptionIDs = append(salesforceSubscriptionIDs,
				f.Salesforce.SubscriptionId)
		case *subscriptionsv1.ListEnterpriseSubscriptionsFilter_InstanceDomain:
			domain, err := subscriptionsv1.NormalizeInstanceDomain(f.InstanceDomain)
			if err != nil {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.Wrap(err, `invalid filter: "domain" provided but invalid`),
				)
			}
			instanceDomains = append(instanceDomains, domain)
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

	subs, err := s.store.ListEnterpriseSubscriptions(
		ctx,
		subscriptions.ListEnterpriseSubscriptionsOptions{
			IDs:                       subscriptionIDs.Values(),
			IsArchived:                isArchived,
			InstanceDomains:           instanceDomains,
			DisplayNameSubstring:      displayNameSubstring,
			SalesforceSubscriptionIDs: salesforceSubscriptionIDs,

			PageSize: int(req.Msg.GetPageSize()),
		},
	)
	if err != nil {
		return nil, connectutil.InternalError(ctx, logger, err, "list subscriptions")
	}

	accessedSubscriptions := map[string]struct{}{}
	protoSubscriptions := make([]*subscriptionsv1.EnterpriseSubscription, len(subs))
	for i, s := range subs {
		protoSubscriptions[i] = convertSubscriptionToProto(s)
		accessedSubscriptions[s.ID] = struct{}{}
	}
	logger.Scoped("audit").Info("ListEnterpriseSubscriptions",
		log.Strings("accessedSubscriptions", maps.Keys(accessedSubscriptions)),
	)

	return connect.NewResponse(
		&subscriptionsv1.ListEnterpriseSubscriptionsResponse{
			// Never a next page, pagination is not implemented yet:
			// https://linear.app/sourcegraph/issue/CORE-134
			NextPageToken: "",
			Subscriptions: protoSubscriptions,
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

	opts := subscriptions.ListLicensesOpts{
		PageSize: int(req.Msg.GetPageSize()),
	}

	// Validate filters
	filters := req.Msg.GetFilters()
	for _, filter := range filters {
		switch f := filter.GetFilter().(type) {
		case *subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_Type:
			if f.Type == 0 {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.New(`invalid filter: "type" is not valid`),
				)
			}
			if opts.LicenseType != 0 {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.New(`invalid filter: "type" provided more than once`),
				)
			}
			opts.LicenseType = f.Type

		case *subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_LicenseKeySubstring:
			const minLength = 3
			if len(f.LicenseKeySubstring) < minLength {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.Newf(`invalid filter: "license_key_substring" must be longer than %d characters`, minLength),
				)
			}
			if opts.LicenseKeySubstring != "" {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.New(`invalid filter: "license_key_substring"" provided multiple times`),
				)
			}
			opts.LicenseKeySubstring = f.LicenseKeySubstring

		case *subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_SubscriptionId:
			if f.SubscriptionId == "" {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.New(`invalid filter: "subscription_id"" provided but is empty`),
				)
			}
			if opts.SubscriptionID != "" {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.New(`invalid filter: "subscription_id"" provided multiple times`),
				)
			}
			opts.SubscriptionID = f.SubscriptionId

		case *subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_SalesforceOpportunityId:
			if f.SalesforceOpportunityId == "" {
				return nil, connect.NewError(
					connect.CodeInvalidArgument,
					errors.New(`invalid filter: "salesforce_opportunity_id" provided but is empty`),
				)
			}
			opts.SalesforceOpportunityID = f.SalesforceOpportunityId
		}
	}

	if opts.LicenseType != subscriptionsv1.EnterpriseSubscriptionLicenseType_ENTERPRISE_SUBSCRIPTION_LICENSE_TYPE_KEY {
		if opts.LicenseKeySubstring != "" {
			return nil, connect.NewError(
				connect.CodeInvalidArgument,
				errors.New(`invalid filters: "license_type" must be 'ENTERPRISE_SUBSCRIPTION_LICENSE_TYPE_KEY' to use the "license_key_substring" filter`),
			)
		}
		if opts.SalesforceOpportunityID != "" {
			return nil, connect.NewError(
				connect.CodeInvalidArgument,
				errors.New(`invalid filters: "license_type" must be 'ENTERPRISE_SUBSCRIPTION_LICENSE_TYPE_KEY' to use the "salesforce_opportunity_id" filter`),
			)
		}
	}

	licenses, err := s.store.ListEnterpriseSubscriptionLicenses(ctx, opts)
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
		resp.Licenses[i], err = convertLicenseToProto(l)
		if err != nil {
			return nil, connectutil.InternalError(ctx, logger, err,
				"failed to read Enterprise Subscription license")
		}
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

	var opts subscriptions.UpsertSubscriptionOptions

	fieldPaths := req.Msg.GetUpdateMask().GetPaths()
	// Empty field paths means update all non-empty fields.
	if len(fieldPaths) == 0 {
		if v := req.Msg.GetSubscription().GetInstanceDomain(); v != "" {
			opts.InstanceDomain = database.NewNullString(v)
		}
		if v := req.Msg.GetSubscription().GetDisplayName(); v != "" {
			opts.DisplayName = database.NewNullString(v)
		}
	} else {
		for _, p := range fieldPaths {
			switch p {
			case "instance_domain":
				opts.InstanceDomain =
					database.NewNullString(req.Msg.GetSubscription().GetInstanceDomain())
			case "display_name":
				opts.DisplayName =
					database.NewNullString(req.Msg.GetSubscription().GetDisplayName())
			case "*":
				opts.ForceUpdate = true
				opts.InstanceDomain =
					database.NewNullString(req.Msg.GetSubscription().GetInstanceDomain())
				opts.DisplayName =
					database.NewNullString(req.Msg.GetSubscription().GetDisplayName())
			default:
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.Newf("unknown field path: %s", p))
			}
		}
	}

	// Validate and normalize the domain
	if opts.InstanceDomain != nil && opts.InstanceDomain.Valid {
		normalizedDomain, err := subscriptionsv1.NormalizeInstanceDomain(opts.InstanceDomain.String)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid instance domain"))
		}
		opts.InstanceDomain.String = normalizedDomain
	}

	subscription, err := s.store.UpsertEnterpriseSubscription(ctx, subscriptionID, opts)
	if err != nil {
		return nil, connectutil.InternalError(ctx, logger, err, "upsert subscription")
	}

	respSubscription := convertSubscriptionToProto(subscription)
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
		// Double check that the subscription ID is valid.
		subscriptionAttrs, err := s.store.ListEnterpriseSubscriptions(ctx, subscriptions.ListEnterpriseSubscriptionsOptions{
			IDs: []string{subscriptionID},
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
			subscriptions.ListEnterpriseSubscriptionsOptions{
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
	seenRoles := map[subscriptionsv1.Role]struct{}{}
	for _, role := range roles {
		if _, ok := seenRoles[role]; ok {
			return nil, connect.NewError(connect.CodeInvalidArgument,
				errors.Newf("duplicate role: %s", role))
		}
		seenRoles[role] = struct{}{}

		// "role for subscription"
		roleObject := convertProtoRoleToIAMTupleObject(role, subscriptionID)

		switch role {
		case subscriptionsv1.Role_ROLE_SUBSCRIPTION_CUSTOMER_ADMIN:
			// Subscription customer admin can:
			//	- View cody analytics of the subscription

			// Make sure the customer_admin role is created for the subscription
			// with all the tuples that the customer_admin role has.
			//
			// TODO: We may need a more robust home for this, if we expand more
			// capabilities of the admin role.
			tk := iam.TupleKey{
				// SUBJECT(subscription customer admin)
				Subject: iam.ToTupleSubjectCustomerAdmin(subscriptionID, iam.TupleRelationMember),
				// can RELATION(view)
				TupleRelation: iam.TupleRelationView,
				// OBJECT(subscription cody analytics)
				Object: iam.ToTupleObject(iam.TupleTypeSubscriptionCodyAnalytics, subscriptionID),
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
				// SUBJECT(SAMS user)
				Subject: iam.ToTupleSubjectUser(samsAccountID),
				// is RELATION(member)
				TupleRelation: iam.TupleRelationMember,
				// of OBJECT(role)
				Object: roleObject,
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

	// Revoke membership to all roles that exist but are not specified in the
	// request.
	deletes := make([]iam.TupleKey, 0)
	for rid := range subscriptionsv1.Role_name {
		role := subscriptionsv1.Role(rid)
		if role == subscriptionsv1.Role_ROLE_UNSPECIFIED {
			continue
		}

		if _, ok := seenRoles[role]; !ok {
			roleObject := convertProtoRoleToIAMTupleObject(role, subscriptionID)
			if roleObject == "" {
				continue // unsupported, continue
			}
			tk := iam.TupleKey{
				// SUBJECT(SAMS user)
				Subject: iam.ToTupleSubjectUser(samsAccountID),
				// is RELATION(member)
				TupleRelation: iam.TupleRelationMember,
				// of OBJECT(role)
				Object: roleObject,
			}
			allowed, err := s.store.IAMCheck(ctx, iam.CheckOptions{TupleKey: tk})
			if err != nil {
				return nil, connectutil.InternalError(ctx, logger, err, "check relation tuple in IAM")
			}
			if allowed {
				// Delete tuple
				deletes = append(deletes, tk)
			}
		}
	}

	err = s.store.IAMWrite(
		ctx,
		iam.WriteOptions{
			Writes:  writes,
			Deletes: deletes,
		},
	)
	if err != nil {
		return nil, connectutil.InternalError(ctx, logger, err, "write relation tuples to IAM")
	}
	return connect.NewResponse(&subscriptionsv1.UpdateEnterpriseSubscriptionMembershipResponse{}), nil
}
