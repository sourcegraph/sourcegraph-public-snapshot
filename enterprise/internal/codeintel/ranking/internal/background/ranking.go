package background

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/lsifstore"
	rankingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func exportRankingGraph(
	ctx context.Context,
	store store.Store,
	lsifstore lsifstore.LsifStore,
	logger log.Logger,
	readBatchSize int,
	writeBatchSize int,
) (_, _, _ int, err error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, 0, nil
	}

	graphKey := rankingshared.GraphKey()

	uploads, err := store.GetUploadsForRanking(ctx, graphKey, "ranking", readBatchSize)
	if err != nil {
		return 0, 0, 0, err
	}

	numDefinitionsInserted := 0
	numReferencesInserted := 0

	for _, upload := range uploads {
		documentPaths := []string{}
		if err := lsifstore.InsertDefinitionsAndReferencesForDocument(ctx, upload, graphKey, writeBatchSize, func(ctx context.Context, upload shared.ExportedUpload, rankingBatchSize int, rankingGraphKey, path string, document *scip.Document) error {
			documentPaths = append(documentPaths, path)
			numDefinitions, numReferences, err := setDefinitionsAndReferencesForUpload(ctx, store, upload, rankingBatchSize, rankingGraphKey, path, document)
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

			return 0, 0, 0, err
		}

		if err := store.InsertInitialPathRanks(ctx, upload.ID, documentPaths, graphKey); err != nil {
			logger.Error(
				"Failed to insert initial path counts",
				log.Int("id", upload.ID),
				log.Int("repoID", upload.RepoID),
				log.String("graphKey", graphKey),
				log.Error(err),
			)

			return 0, 0, 0, err
		}

		logger.Info(
			"Processed upload for ranking graph",
			log.Int("id", upload.ID),
			log.String("repo", upload.Repo),
			log.String("root", upload.Root),
		)
	}

	return len(uploads), numDefinitionsInserted, numReferencesInserted, nil
}

const skipPrefix = "lsif ."

func setDefinitionsAndReferencesForUpload(
	ctx context.Context,
	store store.Store,
	upload shared.ExportedUpload,
	rankingBatchNumber int,
	rankingGraphKey, path string,
	document *scip.Document,
) (int, int, error) {
	seenDefinitions := map[string]struct{}{}
	definitions := []shared.RankingDefinitions{}
	for _, occ := range document.Occurrences {
		if occ.Symbol == "" || scip.IsLocalSymbol(occ.Symbol) || strings.HasPrefix(occ.Symbol, skipPrefix) {
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
		if occ.Symbol == "" || scip.IsLocalSymbol(occ.Symbol) || strings.HasPrefix(occ.Symbol, skipPrefix) {
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
			return 0, 0, err
		}
	}

	if len(references) > 0 {
		if err := store.InsertReferencesForRanking(ctx, rankingGraphKey, rankingBatchNumber, shared.RankingReferences{
			UploadID:    upload.ID,
			SymbolNames: references,
		}); err != nil {
			return 0, 0, err
		}
	}

	return len(definitions), len(references), nil
}

func vacuumStaleDefinitions(ctx context.Context, store store.Store) (int, int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	numDefinitionRecordsScanned, numDefinitionRecordsRemoved, err := store.VacuumStaleDefinitions(ctx, rankingshared.GraphKey())
	return numDefinitionRecordsScanned, numDefinitionRecordsRemoved, err
}

func vacuumStaleReferences(ctx context.Context, store store.Store) (int, int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	numReferenceRecordsScanned, numReferenceRecordsRemoved, err := store.VacuumStaleReferences(ctx, rankingshared.GraphKey())
	return numReferenceRecordsScanned, numReferenceRecordsRemoved, err
}

func vacuumStaleInitialPaths(ctx context.Context, store store.Store) (int, int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	numPathRecordsScanned, numStalePathRecordsDeleted, err := store.VacuumStaleInitialPaths(ctx, rankingshared.GraphKey())
	return numPathRecordsScanned, numStalePathRecordsDeleted, err
}

func vacuumStaleGraphs(ctx context.Context, store store.Store) (int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, nil
	}

	return store.VacuumStaleGraphs(ctx, rankingshared.DerivativeGraphKeyFromTime(time.Now()))
}

func vacuumStaleRanks(ctx context.Context, store store.Store) (int, int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	return store.VacuumStaleRanks(ctx, rankingshared.DerivativeGraphKeyFromTime(time.Now()))
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
		rankingshared.DerivativeGraphKeyFromTime(time.Now()),
		batchSize,
	)
}

func mapInitializerRankingGraph(
	ctx context.Context,
	store store.Store,
	batchSize int,
) (
	numInitialPathsProcessed int,
	numInitialPathRanksInserted int,
	err error,
) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	return store.InsertInitialPathCounts(
		ctx,
		rankingshared.DerivativeGraphKeyFromTime(time.Now()),
		batchSize,
	)
}

func reduceRankingGraph(
	ctx context.Context,
	store store.Store,
	batchSize int,
) (numPathRanksInserted int, numPathCountInputsProcessed int, err error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	numPathRanksInserted, numPathCountInputsProcessed, err = store.InsertPathRanks(
		ctx,
		rankingshared.DerivativeGraphKeyFromTime(time.Now()),
		batchSize,
	)
	if err != nil {
		return numPathCountInputsProcessed, numPathCountInputsProcessed, err
	}

	return numPathRanksInserted, numPathCountInputsProcessed, nil
}
