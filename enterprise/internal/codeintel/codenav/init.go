package codenav

import (
	"context"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/sourcegraph/log"

	backgroundjobs "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/internal/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/internal/store"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetService creates or returns an already-initialized symbols service.
// If the service is not yet initialized, it will use the provided dependencies.
func GetService(
	db database.DB,
	codeIntelDB codeintelshared.CodeIntelDB,
	uploadSvc UploadService,
	gitserver GitserverClient,
) *Service {
	svc, _ := initServiceMemo.Init(serviceDependencies{
		db,
		codeIntelDB,
		uploadSvc,
		gitserver,
	})

	return svc
}

type serviceDependencies struct {
	db          database.DB
	codeIntelDB codeintelshared.CodeIntelDB
	uploadSvc   UploadService
	gitserver   GitserverClient
}

var (
	bucketName                   = env.Get("CODEINTEL_CODENAV_RANKING_BUCKET", "lsif-pagerank-experiments", "The GCS bucket.")
	rankingGraphKey              = env.Get("CODEINTEL_CODENAV_RANKING_GRAPH_KEY", "dev", "An identifier of the graph export. Change to start a new export in the configured bucket.")
	rankingGraphBatchSize        = env.MustGetInt("CODEINTEL_CODENAV_RANKING_GRAPH_BATCH_SIZE", 16, "How many uploads to process at once.")
	rankingGraphDeleteBatchSize  = env.MustGetInt("CODEINTEL_CODENAV_RANKING_GRAPH_DELETE_BATCH_SIZE", 32, "How many stale uploads to delete at once.")
	rankingBucketCredentialsFile = env.Get("CODEINTEL_CODENAV_RANKING_GOOGLE_APPLICATION_CREDENTIALS_FILE", "", "The path to a service account key file with access to GCS.")
)

var initServiceMemo = memo.NewMemoizedConstructorWithArg(func(deps serviceDependencies) (*Service, error) {
	store := store.New(deps.db, scopedContext("store"))
	lsifStore := lsifstore.New(deps.codeIntelDB, scopedContext("lsifstore"))
	backgroundJobs := backgroundjobs.New(
		scopedContext("background"),
	)

	rankingBucket := func() *storage.BucketHandle {
		if rankingBucketCredentialsFile == "" && os.Getenv("ENABLE_EXPERIMENTAL_RANKING") == "" {
			return nil
		}

		var opts []option.ClientOption
		if rankingBucketCredentialsFile != "" {
			opts = append(opts, option.WithCredentialsFile(rankingBucketCredentialsFile))
		}

		client, err := storage.NewClient(context.Background(), opts...)
		if err != nil {
			log.Scoped("codenav", "").Error("failed to create storage client", log.Error(err))
			return nil
		}

		return client.Bucket(bucketName)
	}()

	svc := newService(
		store,
		lsifStore,
		deps.uploadSvc,
		deps.gitserver,
		rankingBucket,
		backgroundJobs,
		scopedContext("service"),
	)

	backgroundJobs.SetService(svc)
	return svc, nil
})

func scopedContext(component string) *observation.Context {
	return observation.ScopedContext("codeintel", "codenav", component)
}
