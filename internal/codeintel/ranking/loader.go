package ranking

import (
	"context"
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *Service) RankLoader(interval time.Duration) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return s.loadRanks(ctx)
	}))
}

func (s *Service) loadRanks(ctx context.Context) (err error) {
	if !envvar.SourcegraphDotComMode() && os.Getenv("ENABLE_EXPERIMENTAL_RANKING") == "" {
		return nil
	}
	if s.resultsBucket == nil {
		s.logger.Warn("No result bucket is configured")
		return nil
	}

	var filenames []string
	objects := s.resultsBucket.Objects(ctx, &storage.Query{
		Prefix: resultsBucketObjectKeyPrefix,
	})
	for {
		attrs, err := objects.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return err
		}

		// Don't read the the enclosing folder or success metadata file
		if attrs.Name == resultsBucketObjectKeyPrefix || attrs.Name == resultsBucketObjectKeyPrefix+"_SUCCESS" {
			continue
		}

		filenames = append(filenames, attrs.Name)
	}

	knownFilenames, err := s.store.HasInputFilename(ctx, rankingGraphKey, filenames)
	if err != nil {
		return err
	}

	knownFilenameMap := map[string]struct{}{}
	for _, filename := range knownFilenames {
		knownFilenameMap[filename] = struct{}{}
	}

	filtered := filenames[:0]
	for _, filename := range filenames {
		if _, ok := knownFilenameMap[filename]; !ok {
			filtered = append(filtered, filename)
		}
	}

	for _, name := range filtered {
		r, err := s.resultsBucket.Object(name).NewReader(ctx)
		if err != nil {
			return err
		}

		ranks := map[api.RepoName]map[string][]float64{}

		lastOffset := 0
		cr := &countingReader{r: r}
		reader := csv.NewReader(cr)
		for {
			line, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					break
				}

				return err
			}

			offset := cr.n
			s.operations.numCSVBytesRead.Add(float64(offset - lastOffset))
			lastOffset = offset

			if len(line) < 3 {
				return errors.Newf("malformed line: %v", line)
			}

			repo := api.RepoName(line[0])
			path := line[1]
			strRank := line[2]

			rank, err := strconv.ParseFloat(strRank, 64)
			if err != nil {
				return err
			}

			if _, ok := ranks[repo]; !ok {
				ranks[repo] = map[string][]float64{}
			}

			ranks[repo][path] = []float64{rank}
		}

		if err := s.store.BulkSetDocumentRanks(ctx, rankingGraphKey, name, ranks); err != nil {
			return err
		}

		s.operations.numCSVFilesProcessed.Add(float64(1))
	}

	return nil
}

func (s *Service) RankMerger(interval time.Duration) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return s.mergeRanks(ctx)
	}))
}

func (s *Service) mergeRanks(ctx context.Context) (err error) {
	if !envvar.SourcegraphDotComMode() && os.Getenv("ENABLE_EXPERIMENTAL_RANKING") == "" {
		return nil
	}

	numRepositoriesUpdated, numInputRowsProcessed, err := s.store.MergeDocumentRanks(ctx, rankingGraphKey, inputFileBatchSize)
	if err != nil {
		return err
	}

	s.operations.numRepositoriesUpdated.Add(float64(numRepositoriesUpdated))
	s.operations.numInputRowsProcessed.Add(float64(numInputRowsProcessed))
	return nil
}

// countingReader is an io.Reader that counts the number of bytes sent
// back to the caller.
type countingReader struct {
	r io.Reader
	n int
}

func (r *countingReader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	r.n += n
	return n, err
}
