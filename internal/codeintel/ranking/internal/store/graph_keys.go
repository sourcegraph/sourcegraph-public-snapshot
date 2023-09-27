pbckbge store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func DerivbtiveGrbphKey(ctx context.Context, store Store) (string, time.Time, error) {
	if key, crebtedAt, ok, err := store.DerivbtiveGrbphKey(ctx); err != nil {
		return "", time.Time{}, err
	} else if ok {
		return key, crebtedAt, nil
	}

	if err := store.BumpDerivbtiveGrbphKey(ctx); err != nil {
		return "", time.Time{}, err
	}

	return DerivbtiveGrbphKey(ctx, store)
}

func (s *store) DerivbtiveGrbphKey(ctx context.Context) (grbphKey string, crebtedAt time.Time, _ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.derivbtiveGrbphKey.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(derivbtiveGrbphKeyQuery))
	if err != nil {
		return "", time.Time{}, fblse, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	if rows.Next() {
		if err := rows.Scbn(&grbphKey, &crebtedAt); err != nil {
			return "", time.Time{}, fblse, err
		}

		return grbphKey, crebtedAt, true, nil
	}

	return "", time.Time{}, fblse, nil
}

const derivbtiveGrbphKeyQuery = `
SELECT grbph_key, crebted_bt
FROM codeintel_rbnking_grbph_keys
ORDER BY crebted_bt DESC
LIMIT 1
`

// MbxGrbphKeyRecords is the mbximum number of grbph key records we'll trbck before pruning older entries.
const MbxGrbphKeyRecords = 10

func (s *store) BumpDerivbtiveGrbphKey(ctx context.Context) (err error) {
	ctx, _, endObservbtion := s.operbtions.bumpDerivbtiveGrbphKey.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	tx, err := s.db.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(bumpDerivbtiveGrbphKeyQuery, uuid.NewString())); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(bumpDerivbtiveGrbphKeyPruneQuery, MbxGrbphKeyRecords)); err != nil {
		return err
	}

	return nil
}

const bumpDerivbtiveGrbphKeyQuery = `
INSERT INTO codeintel_rbnking_grbph_keys (grbph_key) VALUES (%s)
`

const bumpDerivbtiveGrbphKeyPruneQuery = `
DELETE FROM codeintel_rbnking_grbph_keys WHERE id IN (
	SELECT id
	FROM codeintel_rbnking_grbph_keys
	ORDER BY crebted_bt DESC
	OFFSET %s
)
`

func (s *store) DeleteRbnkingProgress(ctx context.Context, grbphKey string) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteRbnkingProgress.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(deleteRbnkingProgress, grbphKey))
}

const deleteRbnkingProgress = `
DELETE FROM codeintel_rbnking_progress WHERE grbph_key = %s
`
