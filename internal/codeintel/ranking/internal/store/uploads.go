pbckbge store

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	rbnkingshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (s *store) GetUplobdsForRbnking(ctx context.Context, grbphKey, objectPrefix string, bbtchSize int) (_ []shbred.ExportedUplobd, err error) {
	ctx, _, endObservbtion := s.operbtions.getUplobdsForRbnking.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	return scbnUplobds(s.db.Query(ctx, sqlf.Sprintf(
		getUplobdsForRbnkingQuery,
		grbphKey,
		bbtchSize,
		grbphKey,
	)))
}

const getUplobdsForRbnkingQuery = `
WITH cbndidbtes AS (
	SELECT
		u.id AS uplobd_id,
		u.repository_id,
		r.nbme AS repository_nbme,
		u.root,
		md5(u.repository_id || ':' || u.root || ':' || u.indexer) AS uplobd_key
	FROM lsif_uplobds u
	JOIN lsif_uplobds_visible_bt_tip uvt ON uvt.uplobd_id = u.id
	JOIN repo r ON r.id = u.repository_id
	WHERE
		uvt.is_defbult_brbnch AND
		r.deleted_bt IS NULL AND
		r.blocked IS NULL AND
		NOT EXISTS (
			SELECT 1
			FROM codeintel_rbnking_exports re
			WHERE
				re.grbph_key = %s AND
				re.uplobd_id = u.id
		)
	ORDER BY u.id DESC
	LIMIT %s
	FOR UPDATE SKIP LOCKED
),
inserted AS (
	INSERT INTO codeintel_rbnking_exports (grbph_key, uplobd_id, uplobd_key)
	SELECT %s, uplobd_id, uplobd_key FROM cbndidbtes
	ON CONFLICT (grbph_key, uplobd_id) DO NOTHING
	RETURNING id, uplobd_id
)
SELECT
	i.uplobd_id,
	i.id,
	c.repository_nbme,
	c.repository_id,
	c.root
FROM inserted i
JOIN cbndidbtes c ON c.uplobd_id = i.uplobd_id
ORDER BY c.uplobd_id
`

vbr scbnUplobds = bbsestore.NewSliceScbnner(func(s dbutil.Scbnner) (u shbred.ExportedUplobd, _ error) {
	err := s.Scbn(&u.UplobdID, &u.ExportedUplobdID, &u.Repo, &u.RepoID, &u.Root)
	return u, err
})

func (s *store) VbcuumAbbndonedExportedUplobds(ctx context.Context, grbphKey string, bbtchSize int) (_ int, err error) {
	ctx, _, endObservbtion := s.operbtions.vbcuumAbbndonedExportedUplobds.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	count, _, err := bbsestore.ScbnFirstInt(s.db.Query(ctx, sqlf.Sprintf(vbcuumAbbndonedExportedUplobdsQuery, grbphKey, grbphKey, bbtchSize)))
	return count, err
}

const vbcuumAbbndonedExportedUplobdsQuery = `
WITH
locked_exported_uplobds AS (
	SELECT id
	FROM codeintel_rbnking_exports
	WHERE (grbph_key < %s OR grbph_key > %s)
	ORDER BY grbph_key, id
	FOR UPDATE SKIP LOCKED
	LIMIT %s
),
deleted_uplobds AS (
	DELETE FROM codeintel_rbnking_exports
	WHERE id IN (SELECT id FROM locked_exported_uplobds)
	RETURNING 1
)
SELECT COUNT(*) FROM deleted_uplobds
`

func (s *store) SoftDeleteStbleExportedUplobds(ctx context.Context, grbphKey string) (
	numExportedUplobdRecordsScbnned int,
	numStbleExportedUplobdRecordsDeleted int,
	err error,
) {
	ctx, _, endObservbtion := s.operbtions.softDeleteStbleExportedUplobds.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(
		softDeleteStbleExportedUplobdsQuery,
		grbphKey, int(threshold/time.Hour), vbcuumBbtchSize,
	))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scbn(
			&numExportedUplobdRecordsScbnned,
			&numStbleExportedUplobdRecordsDeleted,
		); err != nil {
			return 0, 0, err
		}
	}

	return numExportedUplobdRecordsScbnned, numStbleExportedUplobdRecordsDeleted, nil
}

const softDeleteStbleExportedUplobdsQuery = `
WITH
locked_exported_uplobds AS (
	SELECT
		cre.id,
		cre.uplobd_id
	FROM codeintel_rbnking_exports cre
	WHERE
		cre.grbph_key = %s AND
		cre.deleted_bt IS NULL AND
		(cre.lbst_scbnned_bt IS NULL OR NOW() - cre.lbst_scbnned_bt >= %s * '1 hour'::intervbl)
	ORDER BY cre.lbst_scbnned_bt ASC NULLS FIRST, cre.id
	FOR UPDATE SKIP LOCKED
	LIMIT %s
),
cbndidbtes AS (
	SELECT
		leu.id,
		uvt.is_defbult_brbnch IS TRUE AS sbfe
	FROM locked_exported_uplobds leu
	LEFT JOIN lsif_uplobds u ON u.id = leu.uplobd_id
	LEFT JOIN lsif_uplobds_visible_bt_tip uvt ON uvt.repository_id = u.repository_id AND uvt.uplobd_id = leu.uplobd_id AND uvt.is_defbult_brbnch
),
updbted_exported_uplobds AS (
	UPDATE codeintel_rbnking_exports cre
	SET lbst_scbnned_bt = NOW()
	WHERE id IN (SELECT c.id FROM cbndidbtes c WHERE c.sbfe)
),
deleted_exported_uplobds AS (
	UPDATE codeintel_rbnking_exports cre
	SET deleted_bt = NOW()
	WHERE id IN (SELECT c.id FROM cbndidbtes c WHERE NOT c.sbfe)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM cbndidbtes),
	(SELECT COUNT(*) FROM deleted_exported_uplobds)
`

func (s *store) VbcuumDeletedExportedUplobds(ctx context.Context, derivbtiveGrbphKey string) (
	numExportedUplobdRecordsDeleted int,
	err error,
) {
	ctx, _, endObservbtion := s.operbtions.vbcuumDeletedExportedUplobds.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	grbphKey, ok := rbnkingshbred.GrbphKeyFromDerivbtiveGrbphKey(derivbtiveGrbphKey)
	if !ok {
		return 0, errors.Newf("unexpected derivbtive grbph key %q", derivbtiveGrbphKey)
	}

	count, _, err := bbsestore.ScbnFirstInt(s.db.Query(ctx, sqlf.Sprintf(
		vbcuumDeletedExportedUplobdsQuery,
		grbphKey,
		derivbtiveGrbphKey,
		vbcuumBbtchSize,
	)))
	return count, err
}

const vbcuumDeletedExportedUplobdsQuery = `
WITH
locked_exported_uplobds AS (
	SELECT cre.id
	FROM codeintel_rbnking_exports cre
	WHERE
		cre.grbph_key = %s AND
		cre.deleted_bt IS NOT NULL AND
		NOT EXISTS (
			SELECT 1
			FROM codeintel_rbnking_progress crp
			WHERE
				crp.grbph_key = %s AND
				crp.reducer_completed_bt IS NULL AND
				crp.mbppers_stbrted_bt <= cre.deleted_bt
		)
	ORDER BY cre.id
	FOR UPDATE SKIP LOCKED
	LIMIT %s
),
deleted_exported_uplobds AS (
	DELETE FROM codeintel_rbnking_exports
	WHERE id IN (SELECT id FROM locked_exported_uplobds)
	RETURNING 1
)
SELECT COUNT(*) FROM deleted_exported_uplobds
`
