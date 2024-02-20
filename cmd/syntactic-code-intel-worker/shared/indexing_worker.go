package shared

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/job_store"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

func NewIndexingWorker(ctx context.Context,
	observationCtx *observation.Context,
	store job_store.SyntacticIndexingJobStore,
	config IndexingWorkerConfig) *workerutil.Worker[*job_store.SyntacticIndexingJob] {

	name := "syntactic_code_intel_indexing_worker"

	handler := &indexingHandler{
		git:    gitserver.NewClient(name),
		config: config,
	}

	return dbworker.NewWorker[*job_store.SyntacticIndexingJob](ctx, store.DBWorkerStore(), handler, workerutil.WorkerOptions{
		Name:                 name,
		Interval:             config.PollInterval,
		HeartbeatInterval:    10 * time.Second,
		Metrics:              workerutil.NewMetrics(observationCtx, name),
		NumHandlers:          config.Concurrency,
		MaximumRuntimePerJob: config.MaximumRuntimePerJob,
	})

}

type indexingHandler struct {
	git    gitserver.Client
	config IndexingWorkerConfig
}

var _ workerutil.Handler[*job_store.SyntacticIndexingJob] = &indexingHandler{}

func (i indexingHandler) Handle(ctx context.Context, logger log.Logger, record *job_store.SyntacticIndexingJob) error {
	logger.Info("Stub indexing worker handling record",
		log.Int("id", record.ID),
		log.String("repository name", record.RepositoryName),
		log.String("commit", record.Commit))

	tarStream, err := i.git.ArchiveReader(ctx, api.RepoName(record.RepositoryName), gitserver.ArchiveOptions{
		Treeish: string(record.Commit),
		Format:  gitserver.ArchiveFormatTar,
	})

	if err != nil {
		return err
	}

	cliPath := i.config.CliPath

	location := path.Join(os.TempDir(), "index.scip")

	// defer os.Remove(location)

	cliCmd := exec.Command(cliPath, "index", "--tar", "-", "--language", "java", "--out", location)

	inPipe, err := cliCmd.StdinPipe()
	// stdoutPipe, err := cliCmd.StdoutPipe()
	stderrPipe, err := cliCmd.StderrPipe()

	// buf := bufio.NewReader(stdoutPipe)

	// for {
	// 	line, _, err := buf.ReadLine()
	// 	if err != nil
	// 	fmt.Println(string(line))
	// }

	if err != nil {
		return err
	}

	err = cliCmd.Start()

	if err != nil {
		return err
	}

	io.Copy(inPipe, tarStream)

	stderr, err := io.ReadAll(stderrPipe)

	fmt.Println(string(stderr))

	err = cliCmd.Wait()

	if err != nil {
		return err
	}

	fmt.Println(location)

	return nil
}
