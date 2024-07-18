package subscriptions

import (
	"gorm.io/gorm"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/internal/utctime"
)

// Subscription is an Enterprise subscription condition record.
type SubscriptionCondition struct {
	// SubscriptionID is the internal unprefixed UUID of the related subscription.
	SubscriptionID string `gorm:"type:uuid;not null"`
	// Status is the type of status corresponding to this condition, corresponding
	// to the API 'EnterpriseSubscriptionCondition.Status'.
	Status string `gorm:"not null"`
	// Message is a human-readable message associated with the condition.
	Message *string `gorm:"size:256"`
	// TransitionTime is the time at which the condition was created, i.e. when
	// the subscription transitioned into this status.
	TransitionTime utctime.Time `gorm:"not null;default:current_timestamp"`
}

func (s *SubscriptionCondition) TableName() string {
	return "enterprise_portal_subscription_conditions"
}

func (t *SubscriptionCondition) RunCustomMigrations(migrator gorm.Migrator) error {
	// Drop the old, generated conditions -> subscription constraint
	if c := "fk_enterprise_portal_subscription_conditions_subscription"; migrator.HasConstraint(t, c) {
		if err := migrator.DropConstraint(t, c); err != nil {
			return err
		}
	}
	return nil
}
