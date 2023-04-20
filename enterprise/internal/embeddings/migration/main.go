package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sourcegraph/log"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate/entities/models"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	eiauthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func migrate(
	ctx context.Context,
	toMigrate []api.RepoName,
	observationCtx *observation.Context,
	config *embeddings.EmbeddingsUploadStoreConfig,
	client *weaviate.Client,
	log logger,
	repoStore database.RepoStore,
) error {
	uploadStore, err := embeddings.NewEmbeddingsUploadStore(ctx, observationCtx, config)
	if err != nil {
		return err
	}

	// track number of object in current batch
	batchHave := 0
	batchWant := 1024

	batch := client.Batch().ObjectsBatcher()

	doMigrate := func(class string, codeOrText embeddings.EmbeddingIndex, repoID api.RepoID, revision api.CommitID) {
		dim := codeOrText.ColumnDimension
		for i := 0; i < len(codeOrText.RowMetadata); i++ {
			batch.WithObjects(&models.Object{
				Class: class,
				Properties: map[string]interface{}{
					"repo":       repoID,
					"file_name":  codeOrText.RowMetadata[i].FileName,
					"start_line": codeOrText.RowMetadata[i].StartLine,
					"end_line":   codeOrText.RowMetadata[i].EndLine,
					"revision":   revision,
				},
				Vector: codeOrText.Embeddings[i*dim : (i+1)*dim],
			})
			batchHave++

			if batchHave%batchWant == 0 {
				_, err := batch.Do(ctx)
				if err != nil {
					log("batch.Do Error: %s", err)
				}
				batchHave = 0
			}
		}
	}

	for _, repo := range toMigrate {
		log("migrating %s ...", repo)

		r, err := repoStore.GetByName(ctx, repo)
		if err != nil {
			log("failed to get repo id for %s: %s", repo, err)
			continue
		}
		repoID := r.ID

		indexName := embeddings.GetRepoEmbeddingIndexName(repo)
		index, err := embeddings.DownloadRepoEmbeddingIndex(ctx, uploadStore, string(indexName))
		if err != nil {
			log("failed to download embedding index for %s: %s", repo, err)
			continue
		}

		doMigrate("Code", index.CodeIndex, repoID, index.Revision)
		doMigrate("Text", index.TextIndex, repoID, index.Revision)
	}

	if batchHave > 0 {
		_, err := batch.Do(ctx)
		if err != nil {
			log("batch.Do Error: %s", err)
		}
	}

	return nil
}

type logger func(format string, a ...any)

// Compile:
// GOOS=linux GOARCH=amd64 go build -o migrate main.go
//
// Copy to embeddings container:
// kubectl cp -n prod migrate embeddings-5d64d57bc6-j89tk:/home/sourcegraph/migrate -c embeddings
//
// Run:
// ./migrate --weaviate_host weaviate.prod.svc.cluster.local github.com/sourcegraph/sourcegraph
func main() {
	usage := `
Usage: migrate [options] <repo1,repo2,...>
Options:
		--weaviate_host <host> (default: localhost:8080)
`
	conf.Init()

	ctx := context.Background()

	weaviateHost := "localhost:8080"
	flag.StringVar(&weaviateHost, "weaviate_host", weaviateHost, "The host of the weaviate instance to migrate to")
	flag.Parse()

	logStdout := func(format string, a ...any) {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		fmt.Fprintf(flag.CommandLine.Output(), format, a...)
	}

	if len(flag.Args()) != 1 {
		logStdout("Usage of %s:\n\n%s\n\n", os.Args[0], strings.TrimSpace(usage))
		os.Exit(1)
	}

	// parse repos
	repos := strings.Split(flag.Args()[0], ",")
	toMigrate := make([]api.RepoName, 0, len(repos))
	for _, repo := range repos {
		toMigrate = append(toMigrate, api.RepoName(repo))
	}

	weaviateClient, err := weaviate.NewClient(weaviate.Config{
		Host:   weaviateHost,
		Scheme: "http",
	})
	if err != nil {
		panic(err)
	}

	_, err = weaviateClient.Schema().Getter().Do(ctx)
	if err != nil {
		panic(err)
	}
	logStdout("successfully connected to Weaviate")

	storeConfig := &embeddings.EmbeddingsUploadStoreConfig{}
	storeConfig.Load()

	logStdout(fmt.Sprintf("loaded config: %+v", storeConfig))

	observationCtx := observation.NewContext(log.Scoped("weaviate", "migration"))

	// Get repo store.
	logStdout("Initializing frontend DB connection...")
	sqlDB := mustInitializeFrontendDB(observationCtx, logStdout)
	logStdout("NewDB...")
	db := database.NewDB(observationCtx.Logger, sqlDB)
	go setAuthzProviders(ctx, db)
	repoStore := db.Repos()

	err = migrate(ctx, toMigrate, observationCtx, storeConfig, weaviateClient, logStdout, repoStore)
	if err != nil {
		logStdout("Failed to migrate: %s", err)
		os.Exit(1)
	}
}

func mustInitializeFrontendDB(observationCtx *observation.Context, mylog logger) *sql.DB {
	mylog("Getting frontend DB connection string...")
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})

	mylog("Ensuring frontend DB connection...")
	db, err := connections.EnsureNewFrontendDB(observationCtx, dsn, "weaviate_migration")
	if err != nil {
		observationCtx.Logger.Fatal("failed to connect to database", log.Error(err))
	}

	return db
}

// SetAuthzProviders periodically refreshes the global authz providers. This changes the repositories that are visible for reads based on the
// current actor stored in an operation's context, which is likely an internal actor for many of
// the jobs configured in this service. This also enables repository update operations to fetch
// permissions from code hosts.
func setAuthzProviders(ctx context.Context, db database.DB) {
	// authz also relies on UserMappings being setup.
	globals.WatchPermissionsUserMapping()

	for range time.NewTicker(eiauthz.RefreshInterval()).C {
		allowAccessByDefault, authzProviders, _, _, _ := eiauthz.ProvidersFromConfig(ctx, conf.Get(), db.ExternalServices(), db)
		authz.SetProviders(allowAccessByDefault, authzProviders)
	}
}
