package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"sync"
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
	shared.Main(enterpriseInit)
}

var cbOnce sync.Once

func enterpriseInit(db *sql.DB, repoStore repos.Store, cf *httpcli.Factory) {
	cbOnce.Do(func() {
		ctx := context.Background()
		a8nStore := a8n.NewStore(db)

		// Set up syncer
		go func() {
			syncer := &a8n.ChangesetSyncer{
				Store:       a8nStore,
				ReposStore:  repoStore,
				HTTPFactory: cf,
			}
			for {
				err := syncer.Sync(ctx)
				if err != nil {
					log15.Error("Syncing Changesets", "err", err)
				}
				time.Sleep(2 * time.Minute)
			}
		}()

		// Set up expired campaign deletion
		go func() {
			for {
				err := a8nStore.DeleteExpiredCampaignPlans(ctx)
				if err != nil {
					log15.Error("DeleteExpiredCampaignPlans", "error", err)
				}
				time.Sleep(2 * time.Minute)
			}
		}()
	})
}
