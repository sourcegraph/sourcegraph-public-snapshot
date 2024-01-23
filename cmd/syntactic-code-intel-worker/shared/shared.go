package shared

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"net/http"
	"time"

	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/authz/providers"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"

	"github.com/sourcegraph/sourcegraph/internal/honey"

	srp "github.com/sourcegraph/sourcegraph/internal/authz/subrepoperms"
	"github.com/sourcegraph/sourcegraph/internal/codeintel"
	codeintelshared "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/lsifuploadstore"
	// "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"

	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const addr = ":3188"

type OutlineScipIndex struct {
	ID                 int        `json:"id"`
	Commit             string     `json:"commit"`
	QueuedAt           time.Time  `json:"queuedAt"`
	State              string     `json:"state"`
	FailureMessage     *string    `json:"failureMessage"`
	StartedAt          *time.Time `json:"startedAt"`
	FinishedAt         *time.Time `json:"finishedAt"`
	ProcessAfter       *time.Time `json:"processAfter"`
	NumResets          int        `json:"numResets"`
	NumFailures        int        `json:"numFailures"`
	RepositoryID       int        `json:"repositoryId"`
	RepositoryName     string     `json:"repositoryName"`
	Outfile            string     `json:"outfile"`
	Rank               *int       `json:"placeInQueue"`
	AssociatedUploadID *int       `json:"associatedUpload"`
	ShouldReindex      bool       `json:"shouldReindex"`
	EnqueuerUserID     int32      `json:"enqueuerUserID"`
}

func (i OutlineScipIndex) RecordID() int {
	return i.ID
}

func (i OutlineScipIndex) RecordUID() string {
	return strconv.Itoa(i.ID)
}

type handler struct {
	// store           store.Store
	// lsifStore       lsifstore.Store
	// gitserverClient gitserver.Client
	// repoStore       RepoStore
	// workerStore     dbworkerstore.Store[uploadsshared.Upload]
	// uploadStore     uploadstore.Store
	// handleOp        *observation.Operation
	// budgetRemaining int64
	// enableBudget    bool
	// uploadSizeGauge prometheus.Gauge
}

func (h *handler) Handle(ctx context.Context, logger log.Logger, record OutlineScipIndex) (err error) {

	logger.Info("Handling {}")

	return nil

}

func Main(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, config Config) error {
	fmt.Println("yooosa")
	logger := observationCtx.Logger

	// Initialize tracing/metrics
	observationCtx = observation.NewContext(logger, observation.Honeycomb(&honey.Dataset{
		Name: "syntactic-code-intel-worker",
	}))

	if err := keyring.Init(ctx); err != nil {
		return errors.Wrap(err, "initializing keyring")
	}

	fmt.Println(config.CliPath)

	// Connect to databases
	db := database.NewDB(logger, mustInitializeDB(observationCtx))
	codeIntelDB := mustInitializeCodeIntelDB(observationCtx)

	var indexColumnsWithNullRank = []*sqlf.Query{
		sqlf.Sprintf("u.id"),
		sqlf.Sprintf("u.commit"),
		sqlf.Sprintf("u.queued_at"),
		sqlf.Sprintf("u.state"),
		sqlf.Sprintf("u.failure_message"),
		sqlf.Sprintf("u.started_at"),
		sqlf.Sprintf("u.finished_at"),
		sqlf.Sprintf("u.process_after"),
		sqlf.Sprintf("u.num_resets"),
		sqlf.Sprintf("u.num_failures"),
		sqlf.Sprintf("u.repository_id"),
		sqlf.Sprintf(`u.repository_name`),
		sqlf.Sprintf(`u.indexer`),
		sqlf.Sprintf(`u.outfile`),
		sqlf.Sprintf("NULL"),
		// sqlf.Sprintf(`(SELECT MAX(id) FROM lsif_uploads WHERE associated_index_id = u.id) AS associated_upload_id`),
		sqlf.Sprintf(`u.should_reindex`),
		sqlf.Sprintf(`u.enqueuer_user_id`),
	}

	options := dbworkerstore.Options[OutlineScipIndex]{
		Name:              "codeintel_outline_scip_index",
		TableName:         "outline_scip_indexes",
		ViewName:          "lsif_indexes_with_repository_name u",
		ColumnExpressions: indexColumnsWithNullRank,
		// Scan:              dbworkerstore.BuildWorkerScan(scanIndex),
		OrderByExpression: sqlf.Sprintf("(u.enqueuer_user_id > 0) DESC, u.queued_at, u.id"),
		// StalledMaxAge:     stalledIndexMaxAge,
		// MaxNumResets:      indexMaxNumResets,
	}

	metrics := workerutil.NewMetrics(observationCtx, "codeintel_upload_processor", workerutil.WithSampler(func(job workerutil.Record) bool { return true }))

	workerStore := dbworkerstore.New(observationCtx, nil, options)

	handle := &handler{}

	// Migrations may take a while, but after they're done we'll immediately
	// spin up a server and can accept traffic. Inform external clients we'll
	// be ready for traffic.
	ready()

	worker1 := []goroutine.BackgroundRoutine{dbworker.NewWorker[OutlineScipIndex](ctx, workerStore, handle, workerutil.WorkerOptions{
		Name:                 "precise_code_intel_upload_worker",
		Description:          "processes outline code-intel uploads",
		NumHandlers:          config.WorkerConcurrency,
		Interval:             config.WorkerPollInterval,
		HeartbeatInterval:    time.Second,
		Metrics:              metrics,
		MaximumRuntimePerJob: config.MaximumRuntimePerJob,
	})}


	// Initialize sub-repo permissions client
	authz.DefaultSubRepoPermsChecker = srp.NewSubRepoPermsClient(db.SubRepoPerms())

	_, err := codeintel.NewServices(codeintel.ServiceDependencies{
		DB:             db,
		CodeIntelDB:    codeIntelDB,
		ObservationCtx: observationCtx,
	})
	if err != nil {
		return errors.Wrap(err, "creating codeintel services")
	}

	uploadStore, err := lsifuploadstore.New(ctx, observationCtx, config.SCIPUploadStoreConfig)
	if err != nil {
		return errors.Wrap(err, "creating upload store")
	}
	if err := initializeUploadStore(ctx, uploadStore); err != nil {
		return errors.Wrap(err, "initializing upload store")
	}

	// // Initialize worker
	// worker := uploads.NewUploadProcessorJob(
	// 	observationCtx,
	// 	services.UploadsService,
	// 	db,
	// 	uploadStore,
	// 	config.WorkerConcurrency,
	// 	config.WorkerBudget,
	// 	config.WorkerPollInterval,
	// 	config.MaximumRuntimePerJob,
	// )

	// // Initialize health server
	server := httpserver.NewFromAddr(addr, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      httpserver.NewHandler(nil),
	})

	// // Go!
	goroutine.MonitorBackgroundRoutines(ctx, append(worker1, server)...)

	return nil
}

func mustInitializeDB(observationCtx *observation.Context) *sql.DB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	sqlDB, err := connections.EnsureNewFrontendDB(observationCtx, dsn, "syntactic-code-intel-worker")
	if err != nil {
		log.Scoped("init db").Fatal("Failed to connect to frontend database", log.Error(err))
	}

	//
	// START FLAILING

	ctx := context.Background()
	db := database.NewDB(observationCtx.Logger, sqlDB)
	go func() {
		for range time.NewTicker(providers.RefreshInterval()).C {
			allowAccessByDefault, authzProviders, _, _, _ := providers.ProvidersFromConfig(ctx, conf.Get(), db)
			authz.SetProviders(allowAccessByDefault, authzProviders)
		}
	}()

	// END FLAILING
	//

	return sqlDB
}

func mustInitializeCodeIntelDB(observationCtx *observation.Context) codeintelshared.CodeIntelDB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})
	db, err := connections.EnsureNewCodeIntelDB(observationCtx, dsn, "syntactic-code-intel-worker")
	if err != nil {
		log.Scoped("init db").Fatal("Failed to connect to codeintel database", log.Error(err))
	}

	return codeintelshared.NewCodeIntelDB(observationCtx.Logger, db)
}

func initializeUploadStore(ctx context.Context, uploadStore uploadstore.Store) error {
	for {
		if err := uploadStore.Init(ctx); err == nil || !isRequestError(err) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func isRequestError(err error) bool {
	return errors.HasType(err, &smithyhttp.RequestSendError{})
}
