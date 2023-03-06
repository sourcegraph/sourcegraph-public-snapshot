package background

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func exportRankingGraph(
	ctx context.Context,
	store store.Store,
	lsifstore lsifstore.LsifStore,
	metrics *metrics,
	logger log.Logger,
	numRoutines int,
	readBatchSize int,
	writeBatchSize int,
) (err error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return nil
	}

	uploads, err := store.GetUploadsForRanking(ctx, conf.CodeIntelRankingDocumentReferenceCountsGraphKey(), "ranking", readBatchSize)
	if err != nil {
		return err
	}

	p := pool.New().WithContext(ctx)

	sharedUploads := make(chan shared.ExportedUpload, len(uploads))
	for _, upload := range uploads {
		sharedUploads <- upload
	}
	close(sharedUploads)

	graphKey := conf.CodeIntelRankingDocumentReferenceCountsGraphKey()
	for i := 0; i < numRoutines; i++ {
		p.Go(func(ctx context.Context) error {
			for upload := range sharedUploads {
				if err := lsifstore.InsertDefinitionsAndReferencesForDocument(ctx, upload, graphKey, writeBatchSize, func(ctx context.Context, upload shared.ExportedUpload, rankingBatchSize int, rankingGraphKey, path string, document *scip.Document) error {
					return setDefinitionsAndReferencesForUpload(ctx, store, metrics, upload, rankingBatchSize, rankingGraphKey, path, document)
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

				logger.Info(
					"Processed upload for ranking graph",
					log.Int("id", upload.ID),
					log.String("repo", upload.Repo),
					log.String("root", upload.Root),
				)
				metrics.numUploadsRead.Inc()
			}

			return nil
		})
	}

	if err := p.Wait(); err != nil {
		return err
	}

	return nil
}

func setDefinitionsAndReferencesForUpload(
	ctx context.Context,
	store store.Store,
	metrics *metrics,
	upload shared.ExportedUpload,
	rankingBatchNumber int,
	rankingGraphKey, path string,
	document *scip.Document,
) error {
	seenDefinitions := map[string]struct{}{}
	definitions := []shared.RankingDefinitions{}
	for _, occ := range document.Occurrences {
		if occ.Symbol == "" || scip.IsLocalSymbol(occ.Symbol) {
			continue
		}

		if scip.SymbolRole_Definition.Matches(occ) {
			definitions = append(definitions, shared.RankingDefinitions{
				UploadID:     upload.ID,
				SymbolName:   occ.Symbol,
				DocumentPath: filepath.Join(upload.Root, path),
			})
			seenDefinitions[occ.Symbol] = struct{}{}
		}
	}

	references := []string{}
	for _, occ := range document.Occurrences {
		if occ.Symbol == "" || scip.IsLocalSymbol(occ.Symbol) {
			continue
		}

		if _, ok := seenDefinitions[occ.Symbol]; ok {
			continue
		}
		if !scip.SymbolRole_Definition.Matches(occ) {
			references = append(references, occ.Symbol)
		}
	}

	if len(definitions) > 0 {
		if err := store.InsertDefinitionsForRanking(ctx, rankingGraphKey, rankingBatchNumber, definitions); err != nil {
			return err
		}

		metrics.numDefinitionsInserted.Add(float64(len(definitions)))
	}

	if len(references) > 0 {
		if err := store.InsertReferencesForRanking(ctx, rankingGraphKey, rankingBatchNumber, shared.RankingReferences{
			UploadID:    upload.ID,
			SymbolNames: references,
		}); err != nil {
			return err
		}

		metrics.numReferencesInserted.Add(float64(len(references)))
	}

	return nil
}

func vacuumRankingGraph(
	ctx context.Context,
	store store.Store,
	metrics *metrics,
) error {
	numStaleDefinitionRecordsDeleted, numStaleReferenceRecordsDeleted, err := store.VacuumStaleDefinitionsAndReferences(ctx, conf.CodeIntelRankingDocumentReferenceCountsGraphKey())
	if err != nil {
		return err
	}
	metrics.numStaleDefinitionRecordsDeleted.Add(float64(numStaleDefinitionRecordsDeleted))
	metrics.numStaleReferenceRecordsDeleted.Add(float64(numStaleReferenceRecordsDeleted))

	numMetadataRecordsDeleted, numInputRecordsDeleted, err := store.VacuumStaleGraphs(ctx, getCurrentGraphKey(time.Now()))
	if err != nil {
		return err
	}
	metrics.numMetadataRecordsDeleted.Add(float64(numMetadataRecordsDeleted))
	metrics.numInputRecordsDeleted.Add(float64(numInputRecordsDeleted))

	numRankRecordsDeleted, err := store.VacuumStaleRanks(ctx, getCurrentGraphKey(time.Now()))
	if err != nil {
		return err
	}
	metrics.numRankRecordsDeleted.Add(float64(numRankRecordsDeleted))

	return nil
}

func mapRankingGraph(
	ctx context.Context,
	store store.Store,
	batchSize int,
) (numReferenceRecordsProcessed int, numInputsInserted int, err error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	return store.InsertPathCountInputs(
		ctx,
		getCurrentGraphKey(time.Now()),
		batchSize,
	)
}

func reduceRankingGraph(
	ctx context.Context,
	store store.Store,
	batchSize int,
) (numPathRanksInserted float64, numPathCountInputsProcessed float64, err error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	numPathRanksInserted, numPathCountInputsProcessed, err = store.InsertPathRanks(
		ctx,
		getCurrentGraphKey(time.Now()),
		batchSize,
	)
	if err != nil {
		return numPathCountInputsProcessed, numPathCountInputsProcessed, err
	}

	return numPathRanksInserted, numPathCountInputsProcessed, nil
}

// getCurrentGraphKey returns a derivative key from the configured parent used for exports
// as well as the current "bucket" of time containing the current instant. Each bucket of
// time is the same configurable length, packed end-to-end since the Unix epoch.
//
// Constructing a graph key for the mapper and reducer jobs in this way ensures that begin
// a fresh map/reduce job on a periodic cadence (equal to the bucket length). Changing the
// parent graph key will also create a new map/reduce job (without switching buckets).
func getCurrentGraphKey(now time.Time) string {
	return fmt.Sprintf("%s-%s-%d",
		conf.CodeIntelRankingDocumentReferenceCountsGraphKey(),
		conf.CodeIntelRankingDocumentReferenceCountsDerivativeGraphKeyPrefix(),
		now.UTC().Unix()/int64(conf.CodeIntelRankingStaleResultAge().Seconds()),
	)
}
