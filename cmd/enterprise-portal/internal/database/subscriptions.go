package database

// Subscription is a product subscription record.
type Subscription struct {
	// ID is the prefixed UUID-format identifier for the subscription.
	ID string `gorm:"primaryKey"`
	// InstanceDomain is the instance domain associated with the subscription, e.g.
	// "acme.sourcegraphcloud.com".
	InstanceDomain string `gorm:"index"`
}

// SubscriptionMember is a product subscription membership, which defines the
// access control of which members can access to which subscriptions in what
// degree.
type SubscriptionMember struct {
	// SubscriptionID is the ID of the subscription (aka. object).
	SubscriptionID string `gorm:"not null"`
	// SAMSAccountID is the SAMS account ID of the member (aka. subject).
	SAMSAccountID string `gorm:"not null;index"`
	// Namespace is the namespace of the relationship of the member to the
	// subscription.
	Namespace string `gorm:"not null;index"`
	// Relation is the relationship of the member to the subscription.
	Relation string `gorm:"not null"`
}
