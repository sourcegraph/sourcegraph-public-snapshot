package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) Summaries(ctx context.Context) (_ []shared.Summary, err error) {
	ctx, _, endObservation := s.operations.summaries.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return scanSummaries(s.db.Query(ctx, sqlf.Sprintf(summariesQuery)))
}

var scanSummaries = basestore.NewSliceScanner(scanSummary)

func scanSummary(s dbutil.Scanner) (shared.Summary, error) {
	var (
		graphKey                     string
		mappersStartedAt             time.Time
		mapperCompletedAt            *time.Time
		seedMapperCompletedAt        *time.Time
		reducerStartedAt             *time.Time
		reducerCompletedAt           *time.Time
		numPathRecordsTotal          int
		numReferenceRecordsTotal     int
		numCountRecordsTotal         int
		numPathRecordsProcessed      int
		numReferenceRecordsProcessed int
		numCountRecordsProcessed     int
		visibleToZoekt               bool
	)
	if err := s.Scan(
		&graphKey,
		&mappersStartedAt,
		&mapperCompletedAt,
		&seedMapperCompletedAt,
		&reducerStartedAt,
		&reducerCompletedAt,
		&dbutil.NullInt{N: &numPathRecordsTotal},
		&dbutil.NullInt{N: &numReferenceRecordsTotal},
		&dbutil.NullInt{N: &numCountRecordsTotal},
		&dbutil.NullInt{N: &numPathRecordsProcessed},
		&dbutil.NullInt{N: &numReferenceRecordsProcessed},
		&dbutil.NullInt{N: &numCountRecordsProcessed},
		&visibleToZoekt,
	); err != nil {
		return shared.Summary{}, err
	}

	pathMapperProgress := shared.Progress{
		StartedAt:   mappersStartedAt,
		CompletedAt: seedMapperCompletedAt,
		Processed:   numPathRecordsProcessed,
		Total:       numPathRecordsTotal,
	}

	referenceMapperProgress := shared.Progress{
		StartedAt:   mappersStartedAt,
		CompletedAt: mapperCompletedAt,
		Processed:   numReferenceRecordsProcessed,
		Total:       numReferenceRecordsTotal,
	}

	var reducerProgress *shared.Progress
	if reducerStartedAt != nil {
		reducerProgress = &shared.Progress{
			StartedAt:   *reducerStartedAt,
			CompletedAt: reducerCompletedAt,
			Processed:   numCountRecordsProcessed,
			Total:       numCountRecordsTotal,
		}
	}

	return shared.Summary{
		GraphKey:                graphKey,
		VisibleToZoekt:          visibleToZoekt,
		PathMapperProgress:      pathMapperProgress,
		ReferenceMapperProgress: referenceMapperProgress,
		ReducerProgress:         reducerProgress,
	}, nil
}

const summariesQuery = `
SELECT
	p.graph_key,
	p.mappers_started_at,
	p.mapper_completed_at,
	p.seed_mapper_completed_at,
	p.reducer_started_at,
	p.reducer_completed_at,
	p.num_path_records_total,
	p.num_reference_records_total,
	p.num_count_records_total,
	p.num_path_records_processed,
	p.num_reference_records_processed,
	p.num_count_records_processed,
	COALESCE(p.id = (
		SELECT pl.id
		FROM codeintel_ranking_progress pl
		WHERE pl.reducer_completed_at IS NOT NULL
		ORDER BY pl.reducer_completed_at DESC
		LIMIT 1
	), false) AS visible_to_zoekt
FROM codeintel_ranking_progress p
ORDER BY p.mappers_started_at DESC
`
