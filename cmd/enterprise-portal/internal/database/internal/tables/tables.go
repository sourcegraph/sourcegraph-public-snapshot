package tables

import (
	"gorm.io/gorm/schema"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
)

// ⚠️ WARNING: This list is meant to be read-only.
func AllTables() []schema.Tabler {
	return []schema.Tabler{
		&subscriptions.Subscription{},
	}
}
