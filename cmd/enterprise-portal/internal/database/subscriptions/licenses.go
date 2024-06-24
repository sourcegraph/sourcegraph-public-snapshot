package subscriptions

import (
	"time"

	"github.com/jackc/pgtype"
)

type SubscriptionLicense struct {
	// ⚠️ DO NOT USE: This field is only used for creating foreign key constraint.
	Subscription *Subscription `gorm:"foreignKey:SubscriptionID"`

	// SubscriptionID is the internal unprefixed UUID of the related subscription.
	SubscriptionID string `gorm:"type:uuid;not null"`
	// ID is the internal unprefixed UUID of this license.
	ID string `gorm:"type:uuid;primaryKey"`

	// Timestamps representing the latest timestamps of key conditions related
	// to this subscription.
	//
	// Condition transition details are tracked in 'enterprise_portal_subscription_license_conditions'.
	CreatedAt time.Time  `gorm:"not null;default:current_timestamp"`
	ExpiresAt time.Time  `gorm:"not null"` // All license types should be time-bound.
	RevokedAt *time.Time // Null indicates the licnese is not revoked.

	// LicenseKind is the kind of license stored in LicenseData, corresponding
	// to the API 'EnterpriseSubscriptionLicenseType'.
	LicenseKind string `gorm:"not null"`
	// LicenseData is the license data stored in JSON format. It is read-only
	// and generally never queried in conditions - properties that are should
	// be stored at the subscription or license level.
	//
	// Value shapes correspond to API types appropriate for each
	// 'EnterpriseSubscriptionLicenseType'.
	LicenseData pgtype.JSONB `gorm:"type:jsonb"`
}

func (s *SubscriptionLicense) TableName() string {
	return "enterprise_portal_subscription_licenses"
}
