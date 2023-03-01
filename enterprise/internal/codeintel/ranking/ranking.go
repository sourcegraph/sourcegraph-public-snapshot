package ranking

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *Service) ExportRankingGraph(ctx context.Context, numRoutines int, numBatchSize int, rankingJobEnabled bool) (err error) {
	ctx, _, endObservation := s.operations.exportRankingGraph.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if !rankingJobEnabled {
		return nil
	}

	uploads, err := s.store.GetUploadsForRanking(ctx, ConfigInst.DocumentReferenceCountsGraphKey, "ranking", ConfigInst.SymbolExporterReadBatchSize)
	if err != nil {
		return err
	}

	p := pool.New().WithContext(ctx)

	sharedUploads := make(chan shared.ExportedUpload, len(uploads))
	for _, upload := range uploads {
		sharedUploads <- upload
	}
	close(sharedUploads)

	for i := 0; i < numRoutines; i++ {
		p.Go(func(ctx context.Context) error {
			for upload := range sharedUploads {
				if err := s.lsifstore.InsertDefinitionsAndReferencesForDocument(ctx, upload, ConfigInst.DocumentReferenceCountsGraphKey, numBatchSize, s.setDefinitionsAndReferencesForUpload); err != nil {
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

	if err := p.Wait(); err != nil {
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
	numStaleDefinitionRecordsDeleted, numStaleReferenceRecordsDeleted, err := s.store.VacuumStaleDefinitionsAndReferences(ctx, ConfigInst.DocumentReferenceCountsGraphKey)
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

func (s *Service) MapRankingGraph(ctx context.Context, rankingJobEnabled bool) (numReferenceRecordsProcessed int, numInputsInserted int, err error) {
	ctx, _, endObservation := s.operations.mapRankingGraph.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if !rankingJobEnabled {
		return 0, 0, nil
	}

	return s.store.InsertPathCountInputs(
		ctx,
		getCurrentGraphKey(time.Now()),
		ConfigInst.MapReducerBatchSize,
	)
}

func (s *Service) ReduceRankingGraph(ctx context.Context, rankingJobEnabled bool) (numPathRanksInserted float64, numPathCountInputsProcessed float64, err error) {
	ctx, _, endObservation := s.operations.reduceRankingGraph.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if !rankingJobEnabled {
		return 0, 0, nil
	}

	numPathRanksInserted, numPathCountInputsProcessed, err = s.store.InsertPathRanks(
		ctx,
		getCurrentGraphKey(time.Now()),
		ConfigInst.MapReducerBatchSize,
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
	return fmt.Sprintf("%s-%d", ConfigInst.DocumentReferenceCountsGraphKey, now.UTC().Unix()/int64(ConfigInst.MapReducerInterval.Seconds()))
}
