package upgrades

import (
	"context"
	"fmt"
	"database/sql"
)

// Validate that the Sourcegraph databases are in the expected state.
// - All migrations defined on their version have been registered in migration_logs
// - No schema drift against defined version
func main() {
	// Open basestore connection to dbs, check against known migration list for each database
	fmt.Println("Lets gooooo")
}
