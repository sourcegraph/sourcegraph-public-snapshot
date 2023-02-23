package uploads

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/api/iterator"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/group"
)

func (s *Service) ExportRankingGraph(
	ctx context.Context,
	numRankingRoutines int,
	numBatchSize int,
	rankingJobEnabled bool,
) (err error) {
	ctx, _, endObservation := s.operations.exportRankingGraph.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if !rankingJobEnabled {
		return nil
	}

	uploads, err := s.store.GetUploadsForRanking(ctx, rankingGraphKey, "ranking", rankingGraphBatchSize)
	if err != nil {
		return err
	}

	g := group.New().WithContext(ctx)

	sharedUploads := make(chan shared.ExportedUpload, len(uploads))
	for _, upload := range uploads {
		sharedUploads <- upload
	}
	close(sharedUploads)

	for i := 0; i < numRankingRoutines; i++ {
		g.Go(func(ctx context.Context) error {
			for upload := range sharedUploads {
				if err := s.lsifstore.InsertDefinitionsAndReferencesForDocument(ctx, upload, rankingGraphKey, numBatchSize, s.setDefinitionsAndReferencesForUpload); err != nil {
					s.logger.Error(
						"Failed to process upload for ranking graph",
						log.Int("id", upload.ID),
						log.String("repo", upload.Repo),
						log.String("root", upload.Root),
						log.Error(err),
					)

					return err
				}

				s.logger.Info(
					"Processed upload for ranking graph",
					log.Int("id", upload.ID),
					log.String("repo", upload.Repo),
					log.String("root", upload.Root),
				)
				s.operations.numUploadsRead.Inc()
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *Service) setDefinitionsAndReferencesForUpload(
	ctx context.Context,
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
				Repository:   upload.Repo,
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
		if err := s.store.InsertDefinitionsForRanking(ctx, rankingGraphKey, rankingBatchNumber, definitions); err != nil {
			return err
		}

		s.operations.numDefinitionsInserted.Add(float64(len(definitions)))
	}

	if len(references) > 0 {
		if err := s.store.InsertReferencesForRanking(ctx, rankingGraphKey, rankingBatchNumber, shared.RankingReferences{
			UploadID:    upload.ID,
			SymbolNames: references,
		}); err != nil {
			return err
		}

		s.operations.numReferencesInserted.Add(float64(len(references)))
	}

	return nil
}

func (s *Service) VacuumRankingGraph(ctx context.Context) error {
	numStaleDefinitionRecordsDeleted, numStaleReferenceRecordsDeleted, err := s.store.VacuumStaleDefinitionsAndReferences(ctx, rankingGraphKey)
	if err != nil {
		return err
	}
	s.operations.numStaleDefinitionRecordsDeleted.Add(float64(numStaleDefinitionRecordsDeleted))
	s.operations.numStaleReferenceRecordsDeleted.Add(float64(numStaleReferenceRecordsDeleted))

	numMetadataRecordsDeleted, numInputRecordsDeleted, err := s.store.VacuumStaleGraphs(ctx, getCurrentGraphKey(time.Now()))
	if err != nil {
		return err
	}
	s.operations.numMetadataRecordsDeleted.Add(float64(numMetadataRecordsDeleted))
	s.operations.numInputRecordsDeleted.Add(float64(numInputRecordsDeleted))

	return nil
}

func (s *Service) VacuumRankingGraphOld(ctx context.Context) error {
	if s.rankingBucket == nil {
		return nil
	}

	numDeleted, err := s.store.ProcessStaleExportedUploads(ctx, rankingGraphKey, rankingGraphDeleteBatchSize, func(ctx context.Context, objectPrefix string) error {
		if objectPrefix == "" {
			// Special case: we haven't backfilled some data on dotcom yet
			return nil
		}

		objects := s.rankingBucket.Objects(ctx, &storage.Query{
			Prefix: objectPrefix,
		})
		for {
			attrs, err := objects.Next()
			if err != nil {
				if err == iterator.Done {
					break
				}

				return err
			}

			if err := s.rankingBucket.Object(attrs.Name).Delete(ctx); err != nil {
				return err
			}

			s.operations.numBytesDeleted.Add(float64(attrs.Size))
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.operations.numStaleRecordsDeleted.Add(float64(numDeleted))
	return nil
}

func (s *Service) MapRankingGraph(ctx context.Context, numRankingRoutines int, rankingJobEnabled bool) (
	numReferenceRecordsProcessed int,
	numInputsInserted int,
	err error,
) {
	ctx, _, endObservation := s.operations.mapRankingGraph.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if !rankingJobEnabled {
		return 0, 0, nil
	}

	return s.store.InsertPathCountInputs(
		ctx,
		getCurrentGraphKey(time.Now()),
		rankingMapReduceBatchSize,
	)
}

func (s *Service) ReduceRankingGraph(
	ctx context.Context,
	numRankingRoutines int,
	rankingJobEnabled bool,
) (numPathRanksInserted float64, numPathCountInputsProcessed float64, err error) {
	ctx, _, endObservation := s.operations.reduceRankingGraph.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if !rankingJobEnabled {
		return 0, 0, nil
	}

	numPathRanksInserted, numPathCountInputsProcessed, err = s.store.InsertPathRanks(
		ctx,
		getCurrentGraphKey(time.Now()),
		rankingMapReduceBatchSize,
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
	return fmt.Sprintf("%s-%d", rankingGraphKey, now.UTC().Unix()/int64(RankingConfigInst.Interval.Seconds()))
}
