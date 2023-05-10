package exporter

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/lsifstore"
	rankingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/background"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const recordTypeName = "path count inputs"

func NewSymbolExporter(
	observationCtx *observation.Context,
	store store.Store,
	lsifstore lsifstore.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.symbol-exporter"

	return background.NewPipelineJob(context.Background(), background.PipelineOptions{
		Name:        name,
		Description: "Exports SCIP data to ranking definitions and reference tables.",
		Interval:    config.Interval,
		Metrics:     background.NewPipelineMetrics(observationCtx, name, recordTypeName),
		ProcessFunc: func(ctx context.Context) (numRecordsProcessed int, numRecordsAltered background.TaggedCounts, err error) {
			numUploadsScanned, numDefinitionsInserted, numReferencesInserted, err := exportRankingGraph(
				ctx,
				store,
				lsifstore,
				observationCtx.Logger,
				config.ReadBatchSize,
				config.WriteBatchSize,
			)

			m := map[string]int{
				"definitions": numDefinitionsInserted,
				"references":  numReferencesInserted,
			}
			return numUploadsScanned, background.NewMapCount(m), err
		},
	})
}

func exportRankingGraph(
	ctx context.Context,
	baseStore store.Store,
	baseLsifStore lsifstore.Store,
	logger log.Logger,
	readBatchSize int,
	writeBatchSize int,
) (numUploads, numDefinitionsInserted, numReferencesInserted int, _ error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, 0, nil
	}

	err := baseStore.WithTransaction(ctx, func(tx store.Store) error {
		return baseLsifStore.WithTransaction(ctx, func(lsifTx lsifstore.Store) error {
			graphKey := rankingshared.GraphKey()

			uploads, err := tx.GetUploadsForRanking(ctx, graphKey, "ranking", readBatchSize)
			if err != nil {
				return err
			}
			// assignment to outer scope
			numUploads = len(uploads)

			for _, upload := range uploads {
				documentPaths := []string{}
				if err := lsifTx.InsertDefinitionsAndReferencesForDocument(ctx, upload, graphKey, writeBatchSize, func(ctx context.Context, upload uploadsshared.ExportedUpload, rankingBatchSize int, rankingGraphKey, path string, document *scip.Document) error {
					documentPaths = append(documentPaths, path)
					numDefinitions, numReferences, err := setDefinitionsAndReferencesForUpload(ctx, tx, upload, rankingBatchSize, rankingGraphKey, path, document)

					// assignment to outer scope
					numDefinitionsInserted += numDefinitions
					numReferencesInserted += numReferences
					return err
				}); err != nil {
					logger.Error(
						"Failed to process upload for ranking graph",
						log.Int("id", upload.ID),
						log.String("repo", upload.Repo),
						log.String("root", upload.Root),
						log.Error(err),
					)

					return err
				}

				paths := make(chan string, len(documentPaths))
				for _, path := range documentPaths {
					paths <- path
				}
				close(paths)

				if err := tx.InsertInitialPathRanks(ctx, upload.ID, paths, writeBatchSize, graphKey); err != nil {
					logger.Error(
						"Failed to insert initial path counts",
						log.Int("id", upload.ID),
						log.Int("repoID", upload.RepoID),
						log.String("graphKey", graphKey),
						log.Error(err),
					)

					return err
				}

				logger.Info(
					"Processed upload for ranking graph",
					log.Int("id", upload.ID),
					log.String("repo", upload.Repo),
					log.String("root", upload.Root),
				)

			}

			return nil
		})
	})

	return numUploads, numDefinitionsInserted, numReferencesInserted, err
}

const skipPrefix = "lsif ."

func setDefinitionsAndReferencesForUpload(
	ctx context.Context,
	store store.Store,
	upload uploadsshared.ExportedUpload,
	batchSize int,
	rankingGraphKey, path string,
	document *scip.Document,
) (int, int, error) {
	seenDefinitions, err := setDefinitionsForUpload(ctx, store, upload, rankingGraphKey, path, document)
	if err != nil {
		return 0, 0, err
	}

	references := make(chan string)
	referencesCount := 0

	go func() {
		defer close(references)

		for _, occ := range document.Occurrences {
			if occ.Symbol == "" || scip.IsLocalSymbol(occ.Symbol) || strings.HasPrefix(occ.Symbol, skipPrefix) {
				continue
			}

			if _, ok := seenDefinitions[occ.Symbol]; ok {
				continue
			}
			if !scip.SymbolRole_Definition.Matches(occ) {
				references <- occ.Symbol
				referencesCount++
			}
		}
	}()

	if err := store.InsertReferencesForRanking(ctx, rankingGraphKey, batchSize, upload.ID, references); err != nil {
		for range references {
			// Drain channel to ensure it closes
		}

		return 0, 0, err
	}

	return len(seenDefinitions), referencesCount, nil
}

func setDefinitionsForUpload(
	ctx context.Context,
	store store.Store,
	upload uploadsshared.ExportedUpload,
	rankingGraphKey, path string,
	document *scip.Document,
) (map[string]struct{}, error) {
	seenDefinitions := map[string]struct{}{}
	definitions := make(chan shared.RankingDefinitions)

	go func() {
		defer close(definitions)

		for _, occ := range document.Occurrences {
			if occ.Symbol == "" || scip.IsLocalSymbol(occ.Symbol) || strings.HasPrefix(occ.Symbol, skipPrefix) {
				continue
			}

			if scip.SymbolRole_Definition.Matches(occ) {
				definitions <- shared.RankingDefinitions{
					UploadID:     upload.ID,
					SymbolName:   occ.Symbol,
					DocumentPath: filepath.Join(upload.Root, path),
				}
				seenDefinitions[occ.Symbol] = struct{}{}
			}
		}
	}()

	if err := store.InsertDefinitionsForRanking(ctx, rankingGraphKey, definitions); err != nil {
		for range definitions {
			// Drain channel to ensure it closes
		}

		return nil, err
	}

	return seenDefinitions, nil
}
