package subscriptions

import (
	"time"
)

// Subscription is an Enterprise subscription condition record.
type SubscriptionCondition struct {
	// ⚠️ DO NOT USE: This field is only used for creating foreign key constraint.
	Subscription *Subscription `gorm:"foreignKey:SubscriptionID"`

	// SubscriptionID is the internal unprefixed UUID of the related subscription.
	SubscriptionID string `gorm:"type:uuid;not null"`
	// Status is the type of status corresponding to this condition, corresponding
	// to the API 'EnterpriseSubscriptionCondition.Status'.
	Status string `gorm:"not null"`
	// Message is a human-readable message associated with the condition.
	Message *string `gorm:"size:256"`
	// TransitionTime is the time at which the condition was created, i.e. when
	// the subscription transitioned into this status.
	TransitionTime time.Time `gorm:"not null;default:current_timestamp"`
}

func (s *SubscriptionCondition) TableName() string {
	return "enterprise_portal_subscription_conditions"
}
