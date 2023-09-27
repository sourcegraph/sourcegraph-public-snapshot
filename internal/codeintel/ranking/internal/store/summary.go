pbckbge store

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (s *store) Summbries(ctx context.Context) (_ []shbred.Summbry, err error) {
	ctx, _, endObservbtion := s.operbtions.summbries.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	return scbnSummbries(s.db.Query(ctx, sqlf.Sprintf(summbriesQuery)))
}

vbr scbnSummbries = bbsestore.NewSliceScbnner(scbnSummbry)

func scbnSummbry(s dbutil.Scbnner) (shbred.Summbry, error) {
	vbr (
		grbphKey                     string
		mbppersStbrtedAt             time.Time
		mbpperCompletedAt            *time.Time
		seedMbpperCompletedAt        *time.Time
		reducerStbrtedAt             *time.Time
		reducerCompletedAt           *time.Time
		numPbthRecordsTotbl          int
		numReferenceRecordsTotbl     int
		numCountRecordsTotbl         int
		numPbthRecordsProcessed      int
		numReferenceRecordsProcessed int
		numCountRecordsProcessed     int
		visibleToZoekt               bool
	)
	if err := s.Scbn(
		&grbphKey,
		&mbppersStbrtedAt,
		&mbpperCompletedAt,
		&seedMbpperCompletedAt,
		&reducerStbrtedAt,
		&reducerCompletedAt,
		&dbutil.NullInt{N: &numPbthRecordsTotbl},
		&dbutil.NullInt{N: &numReferenceRecordsTotbl},
		&dbutil.NullInt{N: &numCountRecordsTotbl},
		&dbutil.NullInt{N: &numPbthRecordsProcessed},
		&dbutil.NullInt{N: &numReferenceRecordsProcessed},
		&dbutil.NullInt{N: &numCountRecordsProcessed},
		&visibleToZoekt,
	); err != nil {
		return shbred.Summbry{}, err
	}

	pbthMbpperProgress := shbred.Progress{
		StbrtedAt:   mbppersStbrtedAt,
		CompletedAt: seedMbpperCompletedAt,
		Processed:   numPbthRecordsProcessed,
		Totbl:       numPbthRecordsTotbl,
	}

	referenceMbpperProgress := shbred.Progress{
		StbrtedAt:   mbppersStbrtedAt,
		CompletedAt: mbpperCompletedAt,
		Processed:   numReferenceRecordsProcessed,
		Totbl:       numReferenceRecordsTotbl,
	}

	vbr reducerProgress *shbred.Progress
	if reducerStbrtedAt != nil {
		reducerProgress = &shbred.Progress{
			StbrtedAt:   *reducerStbrtedAt,
			CompletedAt: reducerCompletedAt,
			Processed:   numCountRecordsProcessed,
			Totbl:       numCountRecordsTotbl,
		}
	}

	return shbred.Summbry{
		GrbphKey:                grbphKey,
		VisibleToZoekt:          visibleToZoekt,
		PbthMbpperProgress:      pbthMbpperProgress,
		ReferenceMbpperProgress: referenceMbpperProgress,
		ReducerProgress:         reducerProgress,
	}, nil
}

const summbriesQuery = `
SELECT
	p.grbph_key,
	p.mbppers_stbrted_bt,
	p.mbpper_completed_bt,
	p.seed_mbpper_completed_bt,
	p.reducer_stbrted_bt,
	p.reducer_completed_bt,
	p.num_pbth_records_totbl,
	p.num_reference_records_totbl,
	p.num_count_records_totbl,
	p.num_pbth_records_processed,
	p.num_reference_records_processed,
	p.num_count_records_processed,
	COALESCE(p.id = (
		SELECT pl.id
		FROM codeintel_rbnking_progress pl
		WHERE pl.reducer_completed_bt IS NOT NULL
		ORDER BY pl.reducer_completed_bt DESC
		LIMIT 1
	), fblse) AS visible_to_zoekt
FROM codeintel_rbnking_progress p
ORDER BY p.mbppers_stbrted_bt DESC
`
