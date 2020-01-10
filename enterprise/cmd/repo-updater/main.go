package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func main() {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		log.Println("enterprise edition")
	}

	newPreSync := func(db *sql.DB, rs repos.Store, cf *httpcli.Factory) func(context.Context) error {
		syncer := &a8n.ChangesetSyncer{
			Store:       a8n.NewStore(db),
			ReposStore:  rs,
			HTTPFactory: cf,
		}

		return syncer.Sync
	}

	dbInitHook := func(db *sql.DB) {
		ctx := context.Background()
		store := a8n.NewStore(db)

		for {
			err := store.DeleteExpiredCampaignPlans(ctx)
			if err != nil {
				log15.Error("DeleteExpiredCampaignPlans", "error", err)
			}

			time.Sleep(2 * time.Minute)
		}
	}

	shared.Main(newPreSync, dbInitHook)
}
