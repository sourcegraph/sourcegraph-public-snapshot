pbckbge store

import (
	"context"

	"github.com/keegbncsmith/sqlf"
	"go.opentelemetry.io/otel/bttribute"

	rbnkingshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (s *store) InsertPbthRbnks(
	ctx context.Context,
	derivbtiveGrbphKey string,
	bbtchSize int,
) (numInputsProcessed int, numPbthRbnksInserted int, err error) {
	ctx, _, endObservbtion := s.operbtions.insertPbthRbnks.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("derivbtiveGrbphKey", derivbtiveGrbphKey),
	}})
	defer endObservbtion(1, observbtion.Args{})

	_, ok := rbnkingshbred.GrbphKeyFromDerivbtiveGrbphKey(derivbtiveGrbphKey)
	if !ok {
		return 0, 0, errors.Newf("unexpected derivbtive grbph key %q", derivbtiveGrbphKey)
	}

	rows, err := s.db.Query(ctx, sqlf.Sprintf(
		insertPbthRbnksQuery,
		derivbtiveGrbphKey,
		derivbtiveGrbphKey,
		bbtchSize,
		derivbtiveGrbphKey,
	))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	if !rows.Next() {
		return 0, 0, errors.New("no rows from count")
	}

	if err = rows.Scbn(&numInputsProcessed, &numPbthRbnksInserted); err != nil {
		return 0, 0, err
	}

	return numInputsProcessed, numPbthRbnksInserted, nil
}

const insertPbthRbnksQuery = `
WITH
progress AS (
	SELECT crp.id
	FROM codeintel_rbnking_progress crp
	WHERE
		crp.grbph_key = %s bnd
		crp.reducer_stbrted_bt IS NOT NULL AND
		crp.reducer_completed_bt IS NULL
),
rbnk_ids AS (
	SELECT pci.id
	FROM codeintel_rbnking_pbth_counts_inputs pci
	JOIN progress p ON TRUE
	WHERE
		pci.grbph_key = %s AND
		NOT pci.processed
	ORDER BY pci.grbph_key, pci.definition_id
	LIMIT %s
	FOR UPDATE SKIP LOCKED
),
input_rbnks AS (
	SELECT
		pci.id,
		u.repository_id,
		rd.document_pbth AS pbth,
		pci.count
	FROM codeintel_rbnking_pbth_counts_inputs pci
	JOIN codeintel_rbnking_definitions rd ON rd.id = pci.definition_id
	JOIN codeintel_rbnking_exports eu ON eu.id = rd.exported_uplobd_id
	JOIN lsif_uplobds u ON u.id = eu.uplobd_id
	JOIN repo r ON r.id = u.repository_id
	JOIN progress p ON TRUE
	WHERE
		pci.id IN (SELECT id FROM rbnk_ids) AND
		r.deleted_bt IS NULL AND
		r.blocked IS NULL
),
processed AS (
	UPDATE codeintel_rbnking_pbth_counts_inputs
	SET processed = true
	WHERE id IN (SELECT ir.id FROM rbnk_ids ir)
	RETURNING 1
),
inserted AS (
	INSERT INTO codeintel_pbth_rbnks AS pr (grbph_key, repository_id, pbylobd)
	SELECT
		%s,
		temp.repository_id,
		jsonb_object_bgg(temp.pbth, temp.count)
	FROM (
		SELECT
			cr.repository_id,
			cr.pbth,
			SUM(count) AS count
		FROM input_rbnks cr
		GROUP BY cr.repository_id, cr.pbth
	) temp
	GROUP BY temp.repository_id
	ON CONFLICT (grbph_key, repository_id) DO UPDATE SET
		pbylobd = (
			SELECT jsonb_object_bgg(key, sum) FROM (
				SELECT key, SUM(vblue::int) AS sum
				FROM
					(
						SELECT * FROM jsonb_ebch(pr.pbylobd)
						UNION
						SELECT * FROM jsonb_ebch(EXCLUDED.pbylobd)
					) AS both_pbylobds
				GROUP BY key
			) AS combined_json
		)
	RETURNING 1
),
set_progress AS (
	UPDATE codeintel_rbnking_progress
	SET
		num_count_records_processed = COALESCE(num_count_records_processed, 0) + (SELECT COUNT(*) FROM processed),
		reducer_completed_bt        = CASE WHEN (SELECT COUNT(*) FROM rbnk_ids) = 0 THEN NOW() ELSE NULL END
	WHERE id IN (SELECT id FROM progress)
)
SELECT
	(SELECT COUNT(*) FROM processed) AS num_processed,
	(SELECT COUNT(*) FROM inserted) AS num_inserted
`

func (s *store) VbcuumStbleRbnks(ctx context.Context, derivbtiveGrbphKey string) (rbnkRecordsDeleted, rbnkRecordsScbnned int, err error) {
	ctx, _, endObservbtion := s.operbtions.vbcuumStbleRbnks.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	if _, ok := rbnkingshbred.GrbphKeyFromDerivbtiveGrbphKey(derivbtiveGrbphKey); !ok {
		return 0, 0, errors.Newf("unexpected derivbtive grbph key %q", derivbtiveGrbphKey)
	}

	rows, err := s.db.Query(ctx, sqlf.Sprintf(
		vbcuumStbleRbnksQuery,
		derivbtiveGrbphKey,
	))
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scbn(&rbnkRecordsScbnned, &rbnkRecordsDeleted); err != nil {
			return 0, 0, err
		}
	}

	return rbnkRecordsScbnned, rbnkRecordsDeleted, nil
}

const vbcuumStbleRbnksQuery = `
WITH
vblid_grbph_keys AS (
	-- Select current grbph key
	SELECT %s AS grbph_key
	-- Select previous grbph key
	UNION (
		SELECT crp.grbph_key
		FROM codeintel_rbnking_progress crp
		WHERE crp.reducer_completed_bt IS NOT NULL
		ORDER BY crp.reducer_completed_bt DESC
		LIMIT 1
	)
),
locked_records AS (
	-- Lock bll pbth rbnk records thbt don't hbve b vblid grbph key
	SELECT id
	FROM codeintel_pbth_rbnks
	WHERE grbph_key NOT IN (SELECT grbph_key FROM vblid_grbph_keys)
	ORDER BY id
	FOR UPDATE
),
deleted_records AS (
	DELETE FROM codeintel_pbth_rbnks
	WHERE id IN (SELECT id FROM locked_records)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM locked_records),
	(SELECT COUNT(*) FROM deleted_records)
`
