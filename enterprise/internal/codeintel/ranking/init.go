package ranking

import (
	"context"
	"os"

	"cloud.google.com/go/storage"
	"github.com/sourcegraph/log"
	"google.golang.org/api/option"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/symbols"
)

func NewService(
	observationCtx *observation.Context,
	db database.DB,
	uploadSvc *uploads.Service,
	gitserverClient GitserverClient,
) *Service {
	if resultsGraphKey == "" {
		// The codenav default
		resultsGraphKey = "dev"
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

		return client.Bucket(resultsBucketName)
	}()

	return newService(
		scopedContext("service", observationCtx),
		store.New(scopedContext("store", observationCtx), db),
		uploadSvc,
		gitserverClient,
		symbols.DefaultClient,
		conf.DefaultClient(),
		resultsBucket,
	)
}

type serviceDependencies struct {
	db              database.DB
	uploadsService  *uploads.Service
	gitserverClient GitserverClient
	observationCtx  *observation.Context
}

var (
	// TODO - move these into background config
	resultsBucketName             = env.Get("CODEINTEL_RANKING_RESULTS_BUCKET", "lsif-pagerank-experiments", "The GCS bucket.")
	resultsGraphKey               = env.Get("CODEINTEL_RANKING_RESULTS_GRAPH_KEY", "dev", "An identifier of the graph export. Change to start a new import from the configured bucket.")
	resultsObjectKeyPrefix        = env.Get("CODEINTEL_RANKING_RESULTS_OBJECT_KEY_PREFIX", "ranks/", "The object key prefix that holds results of the last PageRank batch job.")
	resultsBucketCredentialsFile  = env.Get("CODEINTEL_RANKING_RESULTS_GOOGLE_APPLICATION_CREDENTIALS_FILE", "", "The path to a service account key file with access to GCS.")
	exportObjectKeyPrefix         = env.Get("CODEINTEL_RANKING_DEVELOPMENT_EXPORT_OBJECT_KEY_PREFIX", "", "The object key prefix that should be used for development exports.")
	developmentExportRepositories = env.Get("CODEINTEL_RANKING_DEVELOPMENT_EXPORT_REPOSITORIES", "github.com/sourcegraph/sourcegraph,github.com/sourcegraph/lsif-go", "Comma-separated list of repositories whose ranks should be exported for development.")

	// Backdoor tuning for dotcom
	mergeBatchSize = env.MustGetInt("CODEINTEL_RANKING_MERGE_BATCH_SIZE", 5000, "")
)

func scopedContext(component string, observationCtx *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "ranking", component, observationCtx)
}

func NewIndexer(observationCtx *observation.Context, service *Service) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewRepositoryIndexer(
			observationCtx,
			service.store,
			service.gitserverClient,
			service.symbolsClient,
			IndexerConfigInst.Interval,
		),
	}
}

func NewPageRankLoader(observationCtx *observation.Context, service *Service) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		background.NewRankLoader(
			observationCtx,
			service.store,
			service.resultsBucket,
			background.RankLoaderConfig{
				ResultsGraphKey:        resultsGraphKey,
				ResultsObjectKeyPrefix: resultsObjectKeyPrefix,
			},
			LoaderConfigInst.LoadInterval,
		),
		background.NewRankMerger(
			observationCtx,
			service.store,
			service.resultsBucket,
			background.RankMergerConfig{
				ResultsGraphKey:               resultsGraphKey,
				MergeBatchSize:                mergeBatchSize,
				ExportObjectKeyPrefix:         exportObjectKeyPrefix,
				DevelopmentExportRepositories: developmentExportRepositories,
			},
			LoaderConfigInst.MergeInterval,
		),
	}
}
