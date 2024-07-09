package fetcher

import (
	"archive/tar"
	"context"
	"io"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RepositoryFetcher interface {
	FetchRepositoryArchive(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) <-chan ParseRequestOrError
}

type repositoryFetcher struct {
	gitserverClient     gitserver.GitserverClient
	operations          *operations
	maxTotalPathsLength int
	maxFileSize         int64
}

type ParseRequest struct {
	Path string
	Data []byte
}

type ParseRequestOrError struct {
	ParseRequest ParseRequest
	Err          error
}

func NewRepositoryFetcher(observationCtx *observation.Context, gitserverClient gitserver.GitserverClient, maxTotalPathsLength int, maxFileSize int64) RepositoryFetcher {
	return &repositoryFetcher{
		gitserverClient:     gitserverClient,
		operations:          newOperations(observationCtx),
		maxTotalPathsLength: maxTotalPathsLength,
		maxFileSize:         maxFileSize,
	}
}

func (f *repositoryFetcher) FetchRepositoryArchive(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) <-chan ParseRequestOrError {
	requestCh := make(chan ParseRequestOrError)

	go func() {
		defer close(requestCh)

		if err := f.fetchRepositoryArchive(ctx, repo, commit, paths, func(request ParseRequest) {
			requestCh <- ParseRequestOrError{ParseRequest: request}
		}); err != nil {
			requestCh <- ParseRequestOrError{Err: err}
		}
	}()

	return requestCh
}

func (f *repositoryFetcher) fetchRepositoryArchive(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string, callback func(request ParseRequest)) (err error) {
	ctx, trace, endObservation := f.operations.fetchRepositoryArchive.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		repo.Attr(),
		commit.Attr(),
		attribute.Int("paths", len(paths)),
	}})
	defer endObservation(1, observation.Args{})

	f.operations.fetching.Inc()
	defer f.operations.fetching.Dec()

	fetchAndRead := func(paths []string) error {
		rc, err := f.gitserverClient.FetchTar(ctx, repo, commit, paths)
		if err != nil {
			return errors.Wrap(err, "gitserverClient.FetchTar")
		}
		defer rc.Close()

		err = readTar(ctx, tar.NewReader(rc), callback, trace, f.maxFileSize)
		if err != nil {
			return errors.Wrap(err, "readTar")
		}

		return nil
	}

	if len(paths) == 0 {
		// Full archive
		return fetchAndRead(nil)
	}

	// Partial archive
	for _, pathBatch := range batchByTotalLength(paths, f.maxTotalPathsLength) {
		err = fetchAndRead(pathBatch)
		if err != nil {
			return err
		}
	}

	return nil
}

// batchByTotalLength returns batches of paths where each batch contains at most maxTotalLength
// characters, except when a single path exceeds the soft max, in which case that long path will be put
// into its own batch.
func batchByTotalLength(paths []string, maxTotalLength int) [][]string {
	batches := [][]string{}
	currentBatch := []string{}
	currentLength := 0

	for _, path := range paths {
		if len(currentBatch) > 0 && currentLength+len(path) > maxTotalLength {
			batches = append(batches, currentBatch)
			currentBatch = []string{}
			currentLength = 0
		}

		currentBatch = append(currentBatch, path)
		currentLength += len(path)
	}

	batches = append(batches, currentBatch)

	return batches
}

func readTar(ctx context.Context, tarReader *tar.Reader, callback func(request ParseRequest), traceLog observation.TraceLogger, maxFileSize int64) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		tarHeader, err := tarReader.Next()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		if tarHeader.FileInfo().IsDir() || tarHeader.Typeflag == tar.TypeXGlobalHeader {
			continue
		}

		if tarHeader.Size > maxFileSize {
			callback(ParseRequest{Path: tarHeader.Name, Data: []byte{}})
			continue
		}

		data := make([]byte, int(tarHeader.Size))
		traceLog.AddEvent("readTar", attribute.String("event", "reading tar file contents"))
		if _, err := io.ReadFull(tarReader, data); err != nil {
			return err
		}
		traceLog.AddEvent("readTar", attribute.Int64("size", tarHeader.Size))
		callback(ParseRequest{Path: tarHeader.Name, Data: data})
	}
}
