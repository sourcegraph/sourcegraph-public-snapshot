package codeintel

import (
	"context"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	cigit "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type BlameJob struct {
	ID              int
	State           string
	FailureMessage  *string
	QueuedAt        time.Time
	StartedAt       *time.Time
	FinishedAt      *time.Time
	ProcessAfter    *time.Time
	NumResets       int
	NumFailures     int
	LastHeartbeatAt time.Time
	ExecutionLogs   []workerutil.ExecutionLogEntry
	WorkerHostname  string
	Cancel          bool

	RepositoryID int
	AbsolutePath string
}

func (j *BlameJob) RecordID() int {
	return j.ID
}

func Enqueue(ctx context.Context, db database.DB, repoID api.RepoID, absolutePath string) (int, error) {
	var id int
	err := db.QueryRowContext(ctx, `
		INSERT INTO own_blame_jobs (
			repository_id,
			absolute_file_path
		)
		VALUES ($1, $2)
		RETURNING id
	`, repoID, absolutePath,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func State(ctx context.Context, db database.DB, id int) (string, error) {
	var state string
	err := db.QueryRowContext(ctx,
		"SELECT j.state FROM own_blame_jobs AS j WHERE j.id = $1", id,
	).Scan(&state)
	if err != nil {
		return "", err
	}
	return state, nil
}

var blameJobColumns = []*sqlf.Query{
	sqlf.Sprintf("own_blame_jobs.id"),
	sqlf.Sprintf("own_blame_jobs.state"),
	sqlf.Sprintf("own_blame_jobs.failure_message"),
	sqlf.Sprintf("own_blame_jobs.queued_at"),
	sqlf.Sprintf("own_blame_jobs.started_at"),
	sqlf.Sprintf("own_blame_jobs.finished_at"),
	sqlf.Sprintf("own_blame_jobs.process_after"),
	sqlf.Sprintf("own_blame_jobs.num_resets"),
	sqlf.Sprintf("own_blame_jobs.num_failures"),
	sqlf.Sprintf("own_blame_jobs.last_heartbeat_at"),
	sqlf.Sprintf("own_blame_jobs.execution_logs"),
	sqlf.Sprintf("own_blame_jobs.worker_hostname"),
	sqlf.Sprintf("own_blame_jobs.cancel"),
	sqlf.Sprintf("own_blame_jobs.repository_id"),
	sqlf.Sprintf("own_blame_jobs.absolute_file_path"),
}

func scanExampleJob(s dbutil.Scanner) (*BlameJob, error) {
	var job BlameJob
	var executionLogs []dbworkerstore.ExecutionLogEntry

	if err := s.Scan(
		&job.ID,
		&job.State,
		&job.FailureMessage,
		&job.QueuedAt,
		&job.StartedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFailures,
		&job.LastHeartbeatAt,
		pq.Array(&executionLogs),
		&job.WorkerHostname,
		&job.Cancel,
		&job.RepositoryID,
		&job.AbsolutePath,
	); err != nil {
		return nil, err
	}

	for _, entry := range executionLogs {
		job.ExecutionLogs = append(job.ExecutionLogs, workerutil.ExecutionLogEntry(entry))
	}

	return &job, nil
}

func makeStore(observationCtx *observation.Context, dbHandle basestore.TransactableHandle) dbworkerstore.Store[*BlameJob] {
	return dbworkerstore.New(observationCtx, dbHandle, dbworkerstore.Options[*BlameJob]{
		Name:              "own_blame_job_worker_store",
		TableName:         "own_blame_jobs",
		ViewName:          "own_blame_jobs_with_repository_name own_blame_jobs",
		ColumnExpressions: blameJobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(scanExampleJob),
		OrderByExpression: sqlf.Sprintf("own_blame_jobs.repository_id, own_blame_jobs.id"),
		MaxNumResets:      5,
		StalledMaxAge:     time.Second * 5,
	})
}

type handler struct {
	nav   *codenav.Service
	gits  gitserver.Client
	cigit *cigit.Client
	db    database.DB
}

var _ workerutil.Handler[*BlameJob] = &handler{}

func (h *handler) Handle(ctx context.Context, logger log.Logger, record *BlameJob) error {
	fmt.Printf("INVOKED, repo=%d path=%q!\n\n", record.RepositoryID, record.AbsolutePath)
	hunkCache, err := codenav.NewHunkCache(100)
	if err != nil {
		return err
	}
	repos, err := h.db.Repos().GetByIDs(ctx, api.RepoID(record.RepositoryID))
	if err != nil {
		return err
	}
	if got, want := len(repos), 1; got != want {
		return errors.Wrapf(err, "%d repositories with id %d found", got, record.RepositoryID)
	}
	r := repos[0]
	commit, err := h.gits.ResolveRevision(ctx, r.Name, "HEAD", gitserver.ResolveRevisionOptions{})
	if err != nil {
		return err
	}
	hunks, err := h.gits.BlameFile(ctx, nil, r.Name, record.AbsolutePath, &gitserver.BlameOptions{
		NewestCommit: commit,
	})
	if err != nil {
		return err
	}
	maxLine := 0
	owners := map[int]string{}
	if len(hunks) == 0 {
		fmt.Println("no hunks")
	}
	for _, h := range hunks {
		fmt.Println("Hunk", h.StartLine, h.Author.Name)
		if h.EndLine > maxLine {
			maxLine = h.EndLine
		}
		for i := h.StartLine; i <= h.EndLine; i++ {
			owners[i] = h.Author.Name
		}
	}
	args := shared.RequestArgs{
		RepositoryID: record.RepositoryID,
		Commit:       string(commit),
		Path:         record.AbsolutePath,
	}
	uploads, err := h.nav.GetClosestDumpsForBlob(ctx, record.RepositoryID, string(commit), record.AbsolutePath, true, "")
	if err != nil {
		return err
	}
	ranges, err := h.nav.GetRanges(ctx, args, codenav.NewRequestState(uploads, nil, h.cigit, r, string(commit), record.AbsolutePath, 10, hunkCache), 1, maxLine)
	if err != nil {
		return err
	}
	if len(ranges) == 0 {
		fmt.Println("NO RANGES FOR", record.AbsolutePath)
	}
	for _, r := range ranges {
		def := false
		for _, d := range r.Definitions {
			// TODO: Need to check how bad this heuristic is.
			if d.TargetRange.Start.Line <= r.Range.Start.Line && r.Range.End.Line <= d.TargetRange.End.Line {
				def = true
				break
			}
		}
		if def {
			fmt.Println("DEF")
			fmt.Println(owners[r.Range.Start.Line])
		}
	}
	return nil
}

func makeWorker(ctx context.Context, workerStore dbworkerstore.Store[*BlameJob], nav *codenav.Service, gits gitserver.Client, cigit *cigit.Client, db database.DB, metrics workerutil.WorkerObservability) *workerutil.Worker[*BlameJob] {
	handler := &handler{nav: nav, gits: gits, db: db, cigit: cigit}

	return dbworker.NewWorker[*BlameJob](ctx, workerStore, handler, workerutil.WorkerOptions{
		Name:              "own_blame_job_worker",
		Interval:          time.Second, // Poll for a job once per second
		NumHandlers:       1,           // Process only one job at a time (per instance)
		HeartbeatInterval: 10 * time.Second,
		Metrics:           metrics,
	})
}

func makeResetter(logger log.Logger, workerStore dbworkerstore.Store[*BlameJob], m dbworker.ResetterMetrics) *dbworker.Resetter[*BlameJob] {
	return dbworker.NewResetter(logger, workerStore, dbworker.ResetterOptions{
		Name:     "own_blame_job_worker_resetter",
		Interval: time.Second * 30, // Check for orphaned jobs every 30 seconds
		Metrics:  m,
	})
}

type ownBlameJob struct{}

// NewBitbucketProjectPermissionsJob creates a new job for applying explicit permissions
// to all the repositories of a Bitbucket Project.
func NewBlameJob() job.Job {
	return &ownBlameJob{}
}

func (j *ownBlameJob) Description() string {
	return "Associates blame information for definitions in a single file with code intel"
}

func (j *ownBlameJob) Config() []env.Config {
	return nil
}

// Routines is called by the worker service to start the worker.
// It returns a list of goroutines that the worker service should start and manage.
func (j *ownBlameJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}
	store := makeStore(observationCtx, db.Handle())

	rootContext := actor.WithInternalActor(context.Background())
	resetterMetrics := dbworker.NewMetrics(observationCtx, "own_blame")

	services, err := codeintel.InitServices(observationCtx)
	if err != nil {
		return nil, err
	}
	var _ *codenav.Service = services.CodenavService
	workerMetrics := workerutil.NewMetrics(observationCtx, "own_blame_processor", workerutil.WithSampler(func(job workerutil.Record) bool {
		return true
	}))
	gits := gitserver.NewClient(db)
	cigit := cigit.New(observationCtx, db)
	return []goroutine.BackgroundRoutine{
		makeWorker(rootContext, store, services.CodenavService, gits, cigit, db, workerMetrics),
		makeResetter(observationCtx.Logger.Scoped("OwnBlameJob.Resetter", ""), store, *resetterMetrics),
	}, nil
}
