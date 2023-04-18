package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sourcegraph/log"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate/entities/models"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func migrate(
	ctx context.Context,
	toMigrate []api.RepoName,
	observationCtx *observation.Context,
	config *embeddings.EmbeddingsUploadStoreConfig,
	client *weaviate.Client,
	log logger,
) error {
	uploadStore, err := embeddings.NewEmbeddingsUploadStore(ctx, observationCtx, config)
	if err != nil {
		return err
	}

	// track number of object in current batch
	batchHave := 0
	batchWant := 512

	batch := client.Batch().ObjectsBatcher()

	doMigrate := func(class string, codeOrText embeddings.EmbeddingIndex, repoName string) {
		dim := codeOrText.ColumnDimension
		for i := 0; i < len(codeOrText.RowMetadata); i++ {
			batch.WithObjects(&models.Object{
				Class: class,
				Properties: map[string]interface{}{
					"repo":       repoName,
					"file_name":  codeOrText.RowMetadata[i].FileName,
					"start_line": codeOrText.RowMetadata[i].StartLine,
					"end_line":   codeOrText.RowMetadata[i].EndLine,
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
		indexName := embeddings.GetRepoEmbeddingIndexName(repo)
		index, err := embeddings.DownloadRepoEmbeddingIndex(ctx, uploadStore, string(indexName))
		if err != nil {
			log("failed to download embedding index for %s: %s", repo, err)
			continue
		}

		repoName := string(index.RepoName)
		doMigrate("Code", index.CodeIndex, repoName)
		doMigrate("Text", index.TextIndex, repoName)
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

	_, err = weaviateClient.Schema().Getter().Do(context.Background())
	if err != nil {
		panic(err)
	}
	logStdout("successfully connected to Weaviate")

	storeConfig := &embeddings.EmbeddingsUploadStoreConfig{}
	storeConfig.Load()

	logStdout(fmt.Sprintf("loaded config: %+v", storeConfig))

	observationCtx := observation.NewContext(log.Scoped("weaviate", "migration"))

	err = migrate(context.Background(), toMigrate, observationCtx, storeConfig, weaviateClient, logStdout)
	if err != nil {
		logStdout("Failed to migrate: %s", err)
		os.Exit(1)
	}
}
