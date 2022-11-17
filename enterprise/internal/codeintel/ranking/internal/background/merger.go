package background

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rankMerger struct {
	store         store.Store
	resultsBucket *storage.BucketHandle
	metrics       *mergerMetrics
}

type RankMergerConfig struct {
	ResultsGraphKey               string
	MergeBatchSize                int
	ExportObjectKeyPrefix         string
	DevelopmentExportRepositories string
}

func NewRankMerger(
	store store.Store,
	resultsBucket *storage.BucketHandle,
	config RankMergerConfig,
	interval time.Duration,
	observationContext *observation.Context,
) goroutine.BackgroundRoutine {
	merger := &rankMerger{
		store:         store,
		resultsBucket: resultsBucket,
		metrics:       newMergerMetrics(observationContext),
	}

	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return merger.mergeRanks(ctx, config)
	}))
}

func (m *rankMerger) mergeRanks(ctx context.Context, config RankMergerConfig) (err error) {
	if !envvar.SourcegraphDotComMode() && os.Getenv("ENABLE_EXPERIMENTAL_RANKING") == "" {
		return nil
	}

	numRepositoriesUpdated, numInputRowsProcessed, err := m.store.MergeDocumentRanks(ctx, config.ResultsGraphKey, config.MergeBatchSize)
	if err != nil {
		return err
	}

	m.metrics.numRepositoriesUpdated.Add(float64(numRepositoriesUpdated))
	m.metrics.numInputRowsProcessed.Add(float64(numInputRowsProcessed))

	if err := m.exportRanksForDevelopment(ctx, config); err != nil {
		return err
	}

	return nil
}

func (m *rankMerger) exportRanksForDevelopment(ctx context.Context, config RankMergerConfig) (err error) {
	if config.ExportObjectKeyPrefix == "" {
		return nil
	}
	for _, repoName := range strings.Split(config.DevelopmentExportRepositories, ",") {
		lastUpdated, payload, err := m.store.ExportRankPayloadFor(ctx, api.RepoName(repoName))
		if err != nil {
			return err
		}
		lastUpdated = lastUpdated.UTC().Truncate(time.Second)

		objectHandle := m.resultsBucket.Object(filepath.Join(config.ExportObjectKeyPrefix, strings.ReplaceAll(repoName, "/", "_")))

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
