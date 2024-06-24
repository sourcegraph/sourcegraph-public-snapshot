package shared

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	codeintelshared "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/lsifuploadstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/jobstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/uploadhandler"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewIndexingWorker(ctx context.Context,
	observationCtx *observation.Context,
	jobStore jobstore.SyntacticIndexingJobStore,
	config IndexingWorkerConfig,
	db database.DB,
	codeintelDB codeintelshared.CodeIntelDB,
) *workerutil.Worker[*jobstore.SyntacticIndexingJob] {

	name := "syntactic_code_intel_indexing_worker"

	gitserverClient := gitserver.NewClient(name)

	uploadsService := uploads.NewService(observationCtx, db, codeintelDB, gitserverClient)
	uploadsDBStore := uploadsService.UploadHandlerStore()
	uploadstoreStore, _ := lsifuploadstore.New(ctx, observationCtx, config.LSIFUploadStoreConfig)

	handler := &indexingHandler{
		Config:           config,
		GitServerClient:  gitserverClient,
		uploadDBStore:    uploadsDBStore,
		uploadstoreStore: uploadstoreStore,
	}

	return dbworker.NewWorker(ctx, jobStore.DBWorkerStore(), handler, workerutil.WorkerOptions{
		Name:                 name,
		Interval:             config.PollInterval,
		HeartbeatInterval:    10 * time.Second,
		Metrics:              workerutil.NewMetrics(observationCtx, name),
		NumHandlers:          config.Concurrency,
		MaximumRuntimePerJob: config.MaximumRuntimePerJob,
	})

}

type indexingHandler struct {
	Config           IndexingWorkerConfig
	GitServerClient  gitserver.Client
	uploadDBStore    uploadhandler.DBStore[uploads.UploadMetadata]
	uploadstoreStore uploadstore.Store
}

var _ workerutil.Handler[*jobstore.SyntacticIndexingJob] = &indexingHandler{}

func (i indexingHandler) Handle(ctx context.Context, logger log.Logger, record *jobstore.SyntacticIndexingJob) error {
	logger.Info("Stub indexing worker handling record",
		log.Int("id", record.ID),
		log.String("repository name", record.RepositoryName),
		log.String("commit", string(record.Commit)))

	tarStream, err := i.GitServerClient.ArchiveReader(
		ctx,
		api.RepoName(record.RepositoryName),
		gitserver.ArchiveOptions{Treeish: string(record.Commit), Format: gitserver.ArchiveFormatTar},
	)
	if err != nil {
		return err
	}

	tempLocation := path.Join(os.TempDir(), fmt.Sprintf("syntactic-index-job_%d-repo_%d-commit_%s.scip", record.ID, record.RepositoryID, record.Commit))
	// defer func() {
	// 	os.Remove(tempLocation)
	// }()

	command := exec.Command(i.Config.CliPath, "index", "tar", "-", "--language", "java", "--out", tempLocation)

	cmdStdinPipe, err := command.StdinPipe()
	if err != nil {
		return err
	}

	// cmdStderrPipe, err := command.StderrPipe()
	// if err != nil {
	// 	return err
	// }

	err = command.Start()
	if err != nil {
		return err
	}

	_, err = io.Copy(cmdStdinPipe, tarStream)
	if err != nil {
		return err
	}

	// stderr, err := io.ReadAll(cmdStderrPipe)
	// if err != nil {
	// 	return err
	// }
	//
	err = command.Wait()
	if err != nil {
		return err
	}

	logger.Info("Syntactic indexing finished",
		log.Int("repository_id", int(record.RepositoryID)),
		log.String("commit", string(record.Commit)),
		log.String("output", tempLocation))

	f, err := os.Open(tempLocation)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
	}()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	fileSize := fi.Size()

	uploadID, err := uploadhandler.SingleUpload(ctx, i.uploadDBStore, i.uploadstoreStore, uploads.UploadMetadata{
		RepositoryID:   int(record.RepositoryID),
		Commit:         string(record.Commit),
		Root:           "",
		Indexer:        "scip-syntax",
		IndexerVersion: "1.0.0",
		ContentType:    "application/x-protobuf+scip",
	}, &fileSize, gzipReader(bufio.NewReader(f)))

	if err != nil {
		return err
	}

	logger.Info("Successfully queued upload",
		log.Int("uploadID", uploadID),
		log.Int64("uncompressed size", fileSize),
		log.Int("repository_id", int(record.RepositoryID)),
		log.String("commit", string(record.Commit)),
	)

	return nil
}

// gzipReader decorates a source reader by gzip compressing its contents.
func gzipReader(source io.Reader) io.Reader {
	r, w := io.Pipe()
	go func() {
		// propagate gzip write errors into new reader
		w.CloseWithError(gzipPipe(source, w))
	}()
	return r
}

// gzipPipe reads uncompressed data from r and writes compressed data to w.
func gzipPipe(r io.Reader, w io.Writer) (err error) {
	gzipWriter := gzip.NewWriter(w)
	defer func() {
		if closeErr := gzipWriter.Close(); closeErr != nil {
			err = errors.Append(err, err)
		}
	}()

	_, err = io.Copy(gzipWriter, r)
	return err
}
