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
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repoupdater"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
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

func enterpriseInit(db *sql.DB, repoStore repos.Store, cf *httpcli.Factory, server *repoupdater.Server) {
	cbOnce.Do(func() {
		ctx := context.Background()
		campaignsStore := campaigns.NewStore(db)

		syncer := &campaigns.ChangesetSyncer{
			Store:       campaignsStore,
			ReposStore:  repoStore,
			HTTPFactory: cf,
		}
		if server != nil {
			server.ChangesetSyncer = syncer
		}

		// Set up syncer
		go syncer.Run()

		// Set up expired campaign deletion
		go func() {
			for {
				err := campaignsStore.DeleteExpiredCampaignPlans(ctx)
				if err != nil {
					log15.Error("DeleteExpiredCampaignPlans", "error", err)
				}
				time.Sleep(2 * time.Minute)
			}
		}()
	})
}
