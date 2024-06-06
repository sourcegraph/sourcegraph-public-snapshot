package database

// Subscription is a product subscription record.
type Subscription struct {
	// ID is the prefixed UUID-format identifier for the subscription.
	ID string `gorm:"primaryKey"`
	// InstanceDomain is the instance domain associated with the subscription, e.g.
	// "acme.sourcegraphcloud.com".
	InstanceDomain string `gorm:"index"`
}
