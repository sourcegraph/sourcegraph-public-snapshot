package subscriptionsservice

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
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
