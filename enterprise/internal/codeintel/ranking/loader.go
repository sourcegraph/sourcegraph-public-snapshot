package ranking

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/group"
)

func (s *Service) RankLoader(interval time.Duration) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return s.loadRanks(ctx)
	}))
}

const (
	pageRankPrecision      = float64(1.0)
	rankInputFileBatchSize = 256
)

func (s *Service) loadRanks(ctx context.Context) (err error) {
	if !envvar.SourcegraphDotComMode() && os.Getenv("ENABLE_EXPERIMENTAL_RANKING") == "" {
		return nil
	}
	if s.resultsBucket == nil {
		s.logger.Warn("No result bucket is configured")
		return nil
	}

	batches := make(chan []string)
	g := group.New().WithContext(ctx)

	g.Go(func(ctx context.Context) error {
		defer close(batches)

		batch := make([]string, 0, rankInputFileBatchSize)
		objects := s.resultsBucket.Objects(ctx, &storage.Query{
			Prefix: resultsObjectKeyPrefix,
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
			if attrs.Name == resultsObjectKeyPrefix || attrs.Name == resultsObjectKeyPrefix+"_SUCCESS" {
				continue
			}

			batch = append(batch, attrs.Name)

			if len(batch) > rankInputFileBatchSize {
				select {
				case batches <- batch:
				case <-ctx.Done():
					return ctx.Err()
				}

				batch = make([]string, 0, rankInputFileBatchSize)
			}
		}

		if len(batch) > 0 {
			select {
			case batches <- batch:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	})

	g.Go(func(ctx context.Context) error {
		for filenames := range batches {
			knownFilenames, err := s.store.HasInputFilename(ctx, resultsGraphKey, filenames)
			if err != nil {
				return err
			}

			knownFilenameMap := map[string]struct{}{}
			for _, filename := range knownFilenames {
				knownFilenameMap[filename] = struct{}{}
			}

			for _, name := range filenames {
				if _, ok := knownFilenameMap[name]; ok {
					continue
				}

				r, err := s.resultsBucket.Object(name).NewReader(ctx)
				if err != nil {
					return err
				}

				ranks := map[api.RepoName]map[string]float64{}

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
						ranks[repo] = map[string]float64{}
					}
					ranks[repo][path] = rank
				}

				if err := s.store.BulkSetDocumentRanks(ctx, resultsGraphKey, name, pageRankPrecision, ranks); err != nil {
					return err
				}

				s.operations.numCSVFilesProcessed.Add(float64(1))
			}
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return err
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

	numRepositoriesUpdated, numInputRowsProcessed, err := s.store.MergeDocumentRanks(ctx, resultsGraphKey, mergeBatchSize)
	if err != nil {
		return err
	}

	s.operations.numRepositoriesUpdated.Add(float64(numRepositoriesUpdated))
	s.operations.numInputRowsProcessed.Add(float64(numInputRowsProcessed))

	if err := s.exportRanksForDevelopment(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Service) exportRanksForDevelopment(ctx context.Context) (err error) {
	if exportObjectKeyPrefix == "" {
		return nil
	}
	for _, repoName := range strings.Split(developmentExportRepositories, ",") {
		lastUpdated, payload, err := s.store.ExportRankPayloadFor(ctx, api.RepoName(repoName))
		if err != nil {
			return err
		}
		lastUpdated = lastUpdated.UTC().Truncate(time.Second)

		objectHandle := s.resultsBucket.Object(filepath.Join(exportObjectKeyPrefix, strings.ReplaceAll(repoName, "/", "_")))

		if attrs, err := objectHandle.Attrs(ctx); err != nil {
			if err != storage.ErrObjectNotExist {
				return err
			}
		} else {
			if !attrs.CustomTime.IsZero() && !attrs.CustomTime.Before(lastUpdated) {
				continue
			}
		}

		w := objectHandle.NewWriter(ctx)
		if _, err := io.Copy(w, bytes.NewReader(payload)); err != nil {
			_ = w.Close()
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}
		if _, err := objectHandle.Update(ctx, storage.ObjectAttrsToUpdate{CustomTime: lastUpdated}); err != nil {
			return err
		}
	}

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
