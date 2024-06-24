package subscriptionsservice

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/iam"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func convertLicenseAttrsToProto(attrs *dotcomdb.LicenseAttributes) *subscriptionsv1.EnterpriseSubscriptionLicense {
	conds := []*subscriptionsv1.EnterpriseSubscriptionLicenseCondition{
		{
			Status:             subscriptionsv1.EnterpriseSubscriptionLicenseCondition_STATUS_CREATED,
			LastTransitionTime: timestamppb.New(attrs.CreatedAt),
		},
	}
	if attrs.RevokedAt != nil {
		conds = append(conds, &subscriptionsv1.EnterpriseSubscriptionLicenseCondition{
			Status:             subscriptionsv1.EnterpriseSubscriptionLicenseCondition_STATUS_REVOKED,
			LastTransitionTime: timestamppb.New(*attrs.RevokedAt),
			Message:            pointers.DerefZero(attrs.RevokeReason),
		})
	}

	return &subscriptionsv1.EnterpriseSubscriptionLicense{
		Id:             subscriptionsv1.EnterpriseSubscriptionLicenseIDPrefix + attrs.ID,
		SubscriptionId: subscriptionsv1.EnterpriseSubscriptionIDPrefix + attrs.SubscriptionID,
		Conditions:     conds,
		License: &subscriptionsv1.EnterpriseSubscriptionLicense_Key{
			Key: &subscriptionsv1.EnterpriseSubscriptionLicenseKey{
				InfoVersion: pointers.DerefZero(attrs.InfoVersion),
				Info: &subscriptionsv1.EnterpriseSubscriptionLicenseKey_Info{
					Tags:                     attrs.Tags,
					UserCount:                pointers.DerefZero(attrs.UserCount),
					ExpireTime:               timestamppb.New(*attrs.ExpiresAt),
					SalesforceSubscriptionId: pointers.DerefZero(attrs.SalesforceSubscriptionID),
					SalesforceOpportunityId:  pointers.DerefZero(attrs.SalesforceOpportunityID),
				},
				LicenseKey: attrs.LicenseKey,
				InstanceId: pointers.DerefZero(attrs.InstanceID),
			},
		},
	}
}

func convertSubscriptionToProto(subscription *subscriptions.Subscription, attrs *dotcomdb.SubscriptionAttributes) *subscriptionsv1.EnterpriseSubscription {
	// Dotcom equivalent missing is surprising, but let's not panic just yet
	if attrs == nil {
		attrs = &dotcomdb.SubscriptionAttributes{
			ID: subscription.ID,
		}
	}
	conds := []*subscriptionsv1.EnterpriseSubscriptionCondition{
		{
			Status:             subscriptionsv1.EnterpriseSubscriptionCondition_STATUS_CREATED,
			LastTransitionTime: timestamppb.New(attrs.CreatedAt),
		},
	}
	if attrs.ArchivedAt != nil {
		conds = append(conds, &subscriptionsv1.EnterpriseSubscriptionCondition{
			Status:             subscriptionsv1.EnterpriseSubscriptionCondition_STATUS_ARCHIVED,
			LastTransitionTime: timestamppb.New(*attrs.ArchivedAt),
		})
	}

	return &subscriptionsv1.EnterpriseSubscription{
		Id:             subscriptionsv1.EnterpriseSubscriptionIDPrefix + attrs.ID,
		Conditions:     conds,
		InstanceDomain: subscription.InstanceDomain,
	}
}

func convertProtoToIAMTupleObjectType(typ subscriptionsv1.PermissionType) iam.TupleType {
	switch typ {
	case subscriptionsv1.PermissionType_PERMISSION_TYPE_SUBSCRIPTION_CODY_ANALYTICS:
		return iam.TupleTypeSubscriptionCodyAnalytics
	default:
		panic("unexpected permission type")
	}
}

func convertProtoToIAMTupleRelation(action subscriptionsv1.PermissionRelation) iam.TupleRelation {
	switch action {
	case subscriptionsv1.PermissionRelation_PERMISSION_RELATION_VIEW:
		return iam.TupleRelationView
	default:
		panic("unexpected permission relation")
	}
}
