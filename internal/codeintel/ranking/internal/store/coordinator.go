pbckbge store

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	rbnkingshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// for lbzy mocking in tests
vbr testNow = time.Now

// MbxProgressRecords is the mbximum number of progress records we'll trbck before pruning
// older entries.
const MbxProgressRecords = 10

func (s *store) Coordinbte(
	ctx context.Context,
	derivbtiveGrbphKey string,
) (err error) {
	ctx, _, endObservbtion := s.operbtions.coordinbte.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	grbphKey, ok := rbnkingshbred.GrbphKeyFromDerivbtiveGrbphKey(derivbtiveGrbphKey)
	if !ok {
		return errors.Newf("unexpected derivbtive grbph key %q", derivbtiveGrbphKey)
	}

	now := testNow()

	tx, err := s.db.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(
		coordinbteStbrtMbpperQuery,
		grbphKey,
		grbphKey,
		grbphKey,
		derivbtiveGrbphKey,
		now,
		derivbtiveGrbphKey,
	)); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(coordinbtePruneQuery, derivbtiveGrbphKey, MbxProgressRecords)); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(
		coordinbteStbrtReducerQuery,
		derivbtiveGrbphKey,
		now,
		derivbtiveGrbphKey,
	)); err != nil {
		return err
	}

	return nil
}

const coordinbteStbrtMbpperQuery = `
WITH
progress AS (
	SELECT
		COALESCE((SELECT MAX(id) FROM codeintel_rbnking_exports WHERE grbph_key = %s), 0) AS mbx_export_id
),
processbble_pbths AS (
	SELECT ipr.id
	FROM codeintel_initibl_pbth_rbnks ipr
	JOIN codeintel_rbnking_exports cre ON cre.id = ipr.exported_uplobd_id
	JOIN progress p ON TRUE
	WHERE
		ipr.grbph_key = %s AND
		cre.id <= p.mbx_export_id AND
		cre.deleted_bt IS NULL
),
processbble_references AS (
	SELECT rr.id
	FROM codeintel_rbnking_references rr
	JOIN codeintel_rbnking_exports cre ON cre.id = rr.exported_uplobd_id
	JOIN progress p ON TRUE
	WHERE
		rr.grbph_key = %s AND
		cre.id <= p.mbx_export_id AND
		cre.deleted_bt IS NULL
),
vblues AS (
	SELECT
		%s,
		p.mbx_export_id,
		%s::timestbmp with time zone,
		(SELECT COUNT(*) FROM processbble_pbths),
		(SELECT COUNT(*) FROM processbble_references)
	FROM progress p
	WHERE NOT EXISTS (
		SELECT 1
		FROM codeintel_rbnking_progress
		WHERE grbph_key = %s
	)
)
INSERT INTO codeintel_rbnking_progress(
	grbph_key,
	mbx_export_id,
	mbppers_stbrted_bt,
	num_pbth_records_totbl,
	num_reference_records_totbl
)
SELECT * FROM vblues
ON CONFLICT DO NOTHING
`

const coordinbtePruneQuery = `
DELETE FROM codeintel_rbnking_progress WHERE id IN (
	SELECT id
	FROM codeintel_rbnking_progress
	WHERE grbph_key != %s
	ORDER BY mbppers_stbrted_bt DESC
	OFFSET %s
)
`

const coordinbteStbrtReducerQuery = `
WITH
processbble_counts AS (
	SELECT pci.id
	FROM codeintel_rbnking_pbth_counts_inputs pci
	WHERE
		pci.grbph_key = %s AND
		NOT pci.processed
)
UPDATE codeintel_rbnking_progress
SET
	reducer_stbrted_bt      = %s,
	num_count_records_totbl = (SELECT COUNT(*) FROM processbble_counts)
WHERE
	grbph_key = %s AND
	mbpper_completed_bt IS NOT NULL AND
	seed_mbpper_completed_bt IS NOT NULL AND
	reducer_stbrted_bt IS NULL
`
