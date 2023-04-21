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
	"github.com/weaviate/weaviate-go-client/v4/weaviate/fault"
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
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type weaviateClient struct {
	client      *weaviate.Client
	log         logger
	repoStore   database.RepoStore
	uploadStore uploadstore.Store
}

func newWeaviateClient(host string, logger logger) (*weaviateClient, error) {
	ctx := context.Background()

	storeConfig := &embeddings.EmbeddingsUploadStoreConfig{}
	storeConfig.Load()

	observationCtx := observation.NewContext(log.Scoped("weaviate", "migration"))

	// Get repo store.
	sqlDB := mustInitializeFrontendDB(observationCtx, logger)
	db := database.NewDB(observationCtx.Logger, sqlDB)
	go setAuthzProviders(ctx, db)

	repoStore := db.Repos()
	client, err := weaviate.NewClient(weaviate.Config{
		Host:   host,
		Scheme: "http",
	})
	if err != nil {
		return nil, err
	}

	_, err = client.Schema().Getter().Do(ctx)
	if err != nil {
		return nil, err
	}
	logger("successfully connected to weaviate")

	uploadStore, err := embeddings.NewEmbeddingsUploadStore(ctx, observationCtx, storeConfig)
	if err != nil {
		return nil, err
	}

	return &weaviateClient{
		client:      client,
		log:         logger,
		repoStore:   repoStore,
		uploadStore: uploadStore,
	}, nil
}

func (w *weaviateClient) migrate(ctx context.Context, toMigrate []api.RepoName) error {
	// track number of object in current batch
	batchHave := 0
	batchWant := 1024

	batch := w.client.Batch().ObjectsBatcher()

	doMigrate := func(class string, codeOrText embeddings.EmbeddingIndex, repoID api.RepoID, revision api.CommitID) {
		created, err := w.createClass(ctx, class)
		if err != nil {
			w.log("failed to create class %s: %s", class, err)
			return
		}
		if !created {
			w.log("class %s already exists -> skipping", class)
			return
		}

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
					w.log("batch.Do Error: %s", err)
				}
				batchHave = 0
			}
		}
	}

	for _, repoName := range toMigrate {
		w.log("migrating %s ...", repoName)

		// get repo id
		r, err := w.repoStore.GetByName(ctx, repoName)
		if err != nil {
			w.log("failed to get repo id for %s: %s", repoName, err)
			continue
		}

		indexName := embeddings.GetRepoEmbeddingIndexName(repoName)
		index, err := embeddings.DownloadRepoEmbeddingIndex(ctx, w.uploadStore, string(indexName))
		if err != nil {
			w.log("failed to download embedding index for %s: %s", repoName, err)
			continue
		}

		codeClass, textClass := getClassNamesForRepoID(r.ID)

		doMigrate(codeClass, index.CodeIndex, r.ID, index.Revision)
		doMigrate(textClass, index.TextIndex, r.ID, index.Revision)
	}

	// flush remaining objects
	if batchHave > 0 {
		_, err := batch.Do(ctx)
		if err != nil {
			w.log("batch.Do Error: %s", err)
		}
	}

	return nil
}

func getClassNamesForRepoID(repoID api.RepoID) (string, string) {
	return fmt.Sprintf("Code_%d", repoID), fmt.Sprintf("Text_%d", repoID)
}

// createClass creates a new class in the weaviate schema. If the class already exists, it returns true.
func (w *weaviateClient) createClass(ctx context.Context, class string) (bool, error) {
	err := w.client.Schema().ClassCreator().WithClass(&models.Class{
		Class: class,
		VectorIndexConfig: map[string]interface{}{
			"vectorCacheMaxObjects": 1000,
		},
		Properties: []*models.Property{
			{
				Name:     "repo",
				DataType: []string{"int"},
			},
			{
				Name:     "start_line",
				DataType: []string{"int"},
			},
			{
				Name:     "end_line",
				DataType: []string{"int"},
			},
			{
				Name:         "revision",
				DataType:     []string{"string"},
				Tokenization: "field",
			},
			{
				Name:         "file_name",
				DataType:     []string{"string"},
				Tokenization: "field",
			},
		},
		Vectorizer: "none",
	}).Do(ctx)

	weaviateClientError := &fault.WeaviateClientError{}
	if err != nil && errors.As(err, &weaviateClientError) {
		if weaviateClientError.StatusCode == 422 {
			return false, nil
		}
		return false, err
	}

	return true, nil
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

	// This is needed to connect to frontend db
	conf.Init()

	// parse repos
	repos := strings.Split(flag.Args()[0], ",")
	toMigrate := make([]api.RepoName, 0, len(repos))
	for _, repo := range repos {
		toMigrate = append(toMigrate, api.RepoName(repo))
	}

	weaviateClient, err := newWeaviateClient(weaviateHost, logStdout)
	if err != nil {
		logStdout("Failed to create weaviate client: %s", err)
		os.Exit(1)
	}

	err = weaviateClient.migrate(ctx, toMigrate)
	if err != nil {
		logStdout("Failed to migrate: %s", err)
		os.Exit(1)
	}
}
