package subscriptions

import "time"

type SubscriptionLicenseCondition struct {
	// ⚠️ DO NOT USE: This field is only used for creating foreign key constraint.
	License *SubscriptionLicense `gorm:"foreignKey:LicenseID"`

	// SubscriptionID is the internal unprefixed UUID of the related license.
	LicenseID string `gorm:"type:uuid;not null"`
	// Status is the type of status corresponding to this condition, corresponding
	// to the API 'EnterpriseSubscriptionLicenseCondition.Status'.
	Status string `gorm:"not null"`
	// Message is a human-readable message associated with the condition.
	Message *string `gorm:"size:256"`
	// TransitionTime is the time at which the condition was created, i.e. when
	// the license transitioned into this status.
	TransitionTime time.Time `gorm:"not null;default:current_timestamp"`
}

func (s *SubscriptionLicenseCondition) TableName() string {
	return "enterprise_portal_subscription_license_conditions"
}
