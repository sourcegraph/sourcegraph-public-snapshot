package bg

import (
	"context"
	// "github.com/inconshreveable/log15"
	// "github.com/sourcegraph/sourcegraph/internal/db"
)

// EncryptAllTables begins the process of calling encryptTable for each table where encryption has been enabled
func EncryptAllTables(ctx context.Context) {
	// err := db.ExternalServices.EncryptTable(ctx)
	// if err != nil {
	// 	log15.Error(`encrypt.all-tables: unable to encrypt external_services. Please report this issue.`, "error", err)
	// }

	// err = db.ExternalAccounts.EncryptTable(ctx)
	// if err != nil {
	// 	log15.Error(`encrypt.all-tables: unable to encrypt user_external_accounts. Please report this issue.`, "error", err)
	// }

	// err = db.SavedSearches.EncryptTable(ctx)
	// if err != nil {
	// 	log15.Error(`encrypt.all-tables: unable to encrypt saved_searches. Please report this issue.`, "error", err)
	// }

	// err = db.EventLogs.EncryptTable(ctx)
	// if err != nil {
	// 	log15.Error(`encrypt.all-tables: unable to encrypt event_logs. Please report this issue.`, "error", err)
	// }

}
