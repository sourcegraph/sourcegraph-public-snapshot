package background

import (
	"context"
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sourcegraph/log"
	"google.golang.org/api/iterator"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/group"
)

type rankLoader struct {
	store         store.Store
	resultsBucket *storage.BucketHandle
	logger        log.Logger
	metrics       *loaderMetrics
}

type RankLoaderConfig struct {
	ResultsGraphKey        string
	ResultsObjectKeyPrefix string
}

func NewRankLoader(
	observationCtx *observation.Context,
	store store.Store,
	resultsBucket *storage.BucketHandle,
	config RankLoaderConfig,
	interval time.Duration,
) goroutine.BackgroundRoutine {
	loader := &rankLoader{
		store:         store,
		resultsBucket: resultsBucket,
		logger:        log.Scoped("codeintel-rank-loader", "Reads rank data from GCS into Postgres"),
		metrics:       newLoaderMetrics(observationCtx),
	}

	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return loader.loadRanks(ctx, config)
	}))
}

const (
	pageRankPrecision      = float64(1.0)
	rankInputFileBatchSize = 256
)

func (l *rankLoader) loadRanks(ctx context.Context, config RankLoaderConfig) (err error) {
	if !envvar.SourcegraphDotComMode() && os.Getenv("ENABLE_EXPERIMENTAL_RANKING") == "" {
		return nil
	}
	if l.resultsBucket == nil {
		l.logger.Warn("No result bucket is configured")
		return nil
	}

	batches := make(chan []string)
	g := group.New().WithContext(ctx)

	g.Go(func(ctx context.Context) error {
		defer close(batches)

		batch := make([]string, 0, rankInputFileBatchSize)
		objects := l.resultsBucket.Objects(ctx, &storage.Query{
			Prefix: config.ResultsObjectKeyPrefix,
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
			if attrs.Name == config.ResultsObjectKeyPrefix || attrs.Name == config.ResultsObjectKeyPrefix+"_SUCCESS" {
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
			knownFilenames, err := l.store.HasInputFilename(ctx, config.ResultsGraphKey, filenames)
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

				r, err := l.resultsBucket.Object(name).NewReader(ctx)
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
					l.metrics.numCSVBytesRead.Add(float64(offset - lastOffset))
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

				if err := l.store.BulkSetDocumentRanks(ctx, config.ResultsGraphKey, name, pageRankPrecision, ranks); err != nil {
					return err
				}

				l.metrics.numCSVFilesProcessed.Add(float64(1))
			}
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return err
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
