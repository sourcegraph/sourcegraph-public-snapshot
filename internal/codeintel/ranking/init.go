package ranking

import (
	"context"
	"os"

	"cloud.google.com/go/storage"
	"github.com/sourcegraph/log"
	"google.golang.org/api/option"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/symbols"
)

// GetService creates or returns an already-initialized ranking service.
// If the service is not yet initialized, it will use the provided dependencies.
func GetService(
	db database.DB,
	uploadSvc *uploads.Service,
	gitserverClient GitserverClient,
) *Service {
	svc, _ := initServiceMemo.Init(serviceDependencies{
		db,
		uploadSvc,
		gitserverClient,
	})

	return svc
}

type serviceDependencies struct {
	db              database.DB
	uploadsService  *uploads.Service
	gitserverClient GitserverClient
}

var (
	bucketName                   = env.Get("CODEINTEL_RANKING_RESULTS_BUCKET", "lsif-pagerank-experiments", "The GCS bucket.")
	resultsBucketObjectKeyPrefix = env.Get("CODEINTEL_RANKING_RESULTS_OBJECT_KEY_PREFIX", "ranks/", "The object key prefix that holds results of the last PageRank batch job.")
	resultsBucketCredentialsFile = env.Get("CODEINTEL_RANKING_RESULTS_GOOGLE_APPLICATION_CREDENTIALS_FILE", "", "The path to a service account key file with access to GCS.")

	// Set in codenav service
	rankingGraphKey = os.Getenv("CODEINTEL_CODENAV_RANKING_GRAPH_KEY")

	// Backdoor tuning for dotcom
	inputFileBatchSize = env.MustGetInt("CODEINTEL_RANKING_RESULTS_INPUT_FILE_BATCH_SIZE", 5000, "")
)

var initServiceMemo = memo.NewMemoizedConstructorWithArg(func(deps serviceDependencies) (*Service, error) {
	if rankingGraphKey == "" {
		// The codenav default
		rankingGraphKey = "dev"
	}

	resultsBucket := func() *storage.BucketHandle {
		if resultsBucketCredentialsFile == "" && os.Getenv("ENABLE_EXPERIMENTAL_RANKING") == "" {
			return nil
		}

		var opts []option.ClientOption
		if resultsBucketCredentialsFile != "" {
			opts = append(opts, option.WithCredentialsFile(resultsBucketCredentialsFile))
		}

		client, err := storage.NewClient(context.Background(), opts...)
		if err != nil {
			log.Scoped("ranking", "").Error("failed to create storage client", log.Error(err))
			return nil
		}

		return client.Bucket(bucketName)
	}()

	return newService(
		store.New(deps.db, scopedContext("store")),
		deps.uploadsService,
		deps.gitserverClient,
		symbols.DefaultClient,
		siteConfigQuerier{},
		resultsBucket,
		scopedContext("service"),
	), nil
})

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "ranking", component)
}
