package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/pkg/a8n"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

func main() {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		log.Println("enterprise edition")
	}

	shared.Main(func(db *sql.DB, rs repos.Store, cf *httpcli.Factory) func(context.Context) error {
		syncer := &a8n.ChangesetSyncer{
			Store:       a8n.NewStore(db),
			ReposStore:  rs,
			HTTPFactory: cf,
		}

		return syncer.Sync
	})
}
