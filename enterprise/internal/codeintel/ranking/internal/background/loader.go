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
	store store.Store,
	resultsBucket *storage.BucketHandle,
	config RankLoaderConfig,
	interval time.Duration,
	observationContext *observation.Context,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return (&rankLoader{
			store:         store,
			resultsBucket: resultsBucket,
			logger:        log.Scoped("", ""), // TODO
			metrics:       newLoaderMetrics(observationContext),
		}).loadRanks(ctx, config)
	}))
}

const pageRankPrecision = float64(1.0)

func (s *rankLoader) loadRanks(ctx context.Context, config RankLoaderConfig) (err error) {
	if !envvar.SourcegraphDotComMode() && os.Getenv("ENABLE_EXPERIMENTAL_RANKING") == "" {
		return nil
	}
	if s.resultsBucket == nil {
		s.logger.Warn("No result bucket is configured")
		return nil
	}

	var filenames []string
	objects := s.resultsBucket.Objects(ctx, &storage.Query{
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

		filenames = append(filenames, attrs.Name)
	}

	knownFilenames, err := s.store.HasInputFilename(ctx, config.ResultsGraphKey, filenames)
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
			s.metrics.numCSVBytesRead.Add(float64(offset - lastOffset))
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

		if err := s.store.BulkSetDocumentRanks(ctx, config.ResultsGraphKey, name, pageRankPrecision, ranks); err != nil {
			return err
		}

		s.metrics.numCSVFilesProcessed.Add(float64(1))
	}

	return nil
}
