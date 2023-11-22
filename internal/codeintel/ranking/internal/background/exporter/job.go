package exporter

import (
	"context"
	"crypto/md5"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/lsifstore"
	rankingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/background"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

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
		Metrics:     background.NewPipelineMetrics(observationCtx, name),
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
						log.Int("id", upload.UploadID),
						log.String("repo", upload.Repo),
						log.String("root", upload.Root),
						log.Error(err),
					)

					return err
				}

				if err := tx.InsertInitialPathRanks(ctx, upload.ExportedUploadID, documentPaths, writeBatchSize, graphKey); err != nil {
					logger.Error(
						"Failed to insert initial path counts",
						log.Int("id", upload.UploadID),
						log.Int("repoID", upload.RepoID),
						log.String("graphKey", graphKey),
						log.Error(err),
					)

					return err
				}

				logger.Info(
					"Processed upload for ranking graph",
					log.Int("id", upload.UploadID),
					log.String("repo", upload.Repo),
					log.String("root", upload.Root),
				)

			}

			return nil
		})
	})

	return numUploads, numDefinitionsInserted, numReferencesInserted, err
}

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

	references := make(chan [16]byte)
	referencesCount := 0

	go func() {
		defer close(references)

		for _, occ := range document.Occurrences {
			if scip.SymbolRole_Definition.Matches(occ) {
				// We've already handled definitions
				continue
			}

			// We've already emitted this symbol as a definition
			if _, ok := seenDefinitions[occ.Symbol]; ok {
				continue
			}

			// Parse and format symbol into an opaque string for ranking calculations
			if checksum, ok := canonicalizeSymbol(occ.Symbol); ok {
				references <- checksum
				referencesCount++
			}
		}
	}()

	if err := store.InsertReferencesForRanking(ctx, rankingGraphKey, batchSize, upload.ExportedUploadID, references); err != nil {
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
	uploadID := upload.UploadID
	exportedUploadID := upload.ExportedUploadID
	documentPath := filepath.Join(upload.Root, path)

	seenDefinitions := map[string]struct{}{}
	definitions := make(chan shared.RankingDefinitions)

	go func() {
		defer close(definitions)

		for _, occ := range document.Occurrences {
			if !scip.SymbolRole_Definition.Matches(occ) {
				// We only care about definitions
				continue
			}

			if _, ok := seenDefinitions[occ.Symbol]; ok {
				// We've already emitted a definition for this symbol/file
				continue
			}

			// Parse and format symbol into an opaque string for ranking calculations
			if checksum, ok := canonicalizeSymbol(occ.Symbol); ok {
				definitions <- shared.RankingDefinitions{
					UploadID:         uploadID,
					ExportedUploadID: exportedUploadID,
					SymbolChecksum:   checksum,
					DocumentPath:     documentPath,
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

const skipPrefix = "lsif ."

var emptyChecksum = [16]byte{}

// canonicalizeSymbol transforms a symbol name into an opaque string that
// can be matched internally by the ranking machinery.
//
// Canonicalization of a symbol name for ranking makes two transformations:
//
//   - The package version is removed so that we don't need to match SCIP
//     uploads exactly to get a reference count.
//   - We then hash the simplified symbol name into a fixed-sized block that
//     can be matched in constant time against other symbols in Postgres.
func canonicalizeSymbol(symbolName string) ([16]byte, bool) {
	if symbolName == "" || scip.IsLocalSymbol(symbolName) || strings.HasPrefix(symbolName, skipPrefix) {
		return emptyChecksum, false
	}

	symbol, err := noVersionFormatter.Format(symbolName)
	if err != nil {
		return emptyChecksum, false
	}

	return md5.Sum([]byte(symbol)), true
}

var noVersionFormatter = scip.SymbolFormatter{
	OnError:               func(err error) error { return err },
	IncludeScheme:         func(_ string) bool { return true },
	IncludePackageManager: func(_ string) bool { return true },
	IncludePackageName:    func(_ string) bool { return true },
	IncludePackageVersion: func(_ string) bool { return false },
	IncludeDescriptor:     func(_ string) bool { return true },
	IncludeRawDescriptor:  func(_ *scip.Descriptor) bool { return true },
	IncludeDisambiguator:  func(_ string) bool { return true },
}
