package bg

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func MigrateAllSavedQueriesFromSettingsToDatabase(ctx context.Context) {
	settings, err := db.Settings.ListAll(ctx, "search.savedQueries")
	if err != nil {
		log15.Error(`Warning: unable to migrate "saved queries" to database). Please report this issue`, err)
	}
	for _, s := range settings {
		fmt.Println("All search.savedQueries settings ", *s)
	}

}

// List all saved queries from db using db.Settings.ListAll(ctx, "search.savedQueries")
