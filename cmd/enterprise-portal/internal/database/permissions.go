package database

import (
	"time"
)

// Permission is a Zanzibar-inspired permission record.
type Permission struct {
	// Namespace is the namespace of the permission, e.g. "cody_analytics".
	Namespace string `gorm:"not null;index"`
	// Subject is the subject of the permission, e.g. "User:<SAMS account ID>".
	Subject string `gorm:"not null;index"`
	// Object is the object of the permission, e.g. "Subscription:<subscription ID>".
	Object string `gorm:"not null"`
	// Relation is the relationship between the subject and the object, e.g.
	// "customer_admin".
	Relation string `gorm:"not null"`
	// CommitTime is the time when the permission was committed.
	CommitTime time.Time `gorm:"not null"`
}
