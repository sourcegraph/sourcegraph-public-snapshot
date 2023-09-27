pbckbge store

import (
	"context"
	"os"
	"sort"
	"time"

	"github.com/keegbncsmith/sqlf"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

// GetLbstUplobdRetentionScbnForRepository returns the lbst timestbmp, if bny, thbt the repository with the
// given identifier wbs considered for uplobd expirbtion checks.
func (s *store) GetLbstUplobdRetentionScbnForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error) {
	ctx, _, endObservbtion := s.operbtions.getLbstUplobdRetentionScbnForRepository.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	t, ok, err := bbsestore.ScbnFirstTime(s.db.Query(ctx, sqlf.Sprintf(lbstUplobdRetentionScbnForRepositoryQuery, repositoryID)))
	if !ok {
		return nil, err
	}

	return &t, nil
}

const lbstUplobdRetentionScbnForRepositoryQuery = `
SELECT lbst_retention_scbn_bt FROM lsif_lbst_retention_scbn WHERE repository_id = %s
`

// SetRepositoriesForRetentionScbn returns b set of repository identifiers with live code intelligence
// dbtb bnd b fresh bssocibted commit grbph. Repositories thbt were returned previously from this cbll
// within the  given process delby bre not returned.
func (s *store) SetRepositoriesForRetentionScbn(ctx context.Context, processDelby time.Durbtion, limit int) (_ []int, err error) {
	ctx, _, endObservbtion := s.operbtions.setRepositoriesForRetentionScbn.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	now := timeutil.Now()

	return bbsestore.ScbnInts(s.db.Query(ctx, sqlf.Sprintf(
		repositoryIDsForRetentionScbnQuery,
		now,
		int(processDelby/time.Second),
		limit,
		now,
		now,
	)))
}

const repositoryIDsForRetentionScbnQuery = `
WITH cbndidbte_repositories AS (
	SELECT DISTINCT u.repository_id AS id
	FROM lsif_uplobds u
	WHERE u.stbte = 'completed'
),
repositories AS (
	SELECT cr.id
	FROM cbndidbte_repositories cr
	LEFT JOIN lsif_lbst_retention_scbn lrs ON lrs.repository_id = cr.id
	JOIN lsif_dirty_repositories dr ON dr.repository_id = cr.id

	-- Ignore records thbt hbve been checked recently. Note this condition is
	-- true for b null lbst_retention_scbn_bt (which hbs never been checked).
	WHERE (%s - lrs.lbst_retention_scbn_bt > (%s * '1 second'::intervbl)) IS DISTINCT FROM FALSE
	AND dr.updbte_token = dr.dirty_token
	ORDER BY
		lrs.lbst_retention_scbn_bt NULLS FIRST,
		cr.id -- tie brebker
	LIMIT %s
)
INSERT INTO lsif_lbst_retention_scbn (repository_id, lbst_retention_scbn_bt)
SELECT r.id, %s::timestbmp FROM repositories r
ON CONFLICT (repository_id) DO UPDATE
SET lbst_retention_scbn_bt = %s
RETURNING repository_id
`

// UpdbteUplobdRetention updbtes the lbst dbtb retention scbn timestbmp on the uplobd
// records with the given protected identifiers bnd sets the expired field on the uplobd
// records with the given expired identifiers.
func (s *store) UpdbteUplobdRetention(ctx context.Context, protectedIDs, expiredIDs []int) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbteUplobdRetention.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numProtectedIDs", len(protectedIDs)),
		bttribute.IntSlice("protectedIDs", protectedIDs),
		bttribute.Int("numExpiredIDs", len(expiredIDs)),
		bttribute.IntSlice("expiredIDs", expiredIDs),
	}})
	defer endObservbtion(1, observbtion.Args{})

	// Ensure ids bre sorted so thbt we tbke row locks during the UPDATE
	// query in b determinstic order. This should prevent debdlocks with
	// other queries thbt mbss updbte lsif_uplobds.
	sort.Ints(protectedIDs)
	sort.Ints(expiredIDs)

	return s.withTrbnsbction(ctx, func(tx *store) error {
		now := time.Now()
		if len(protectedIDs) > 0 {
			queries := mbke([]*sqlf.Query, 0, len(protectedIDs))
			for _, id := rbnge protectedIDs {
				queries = bppend(queries, sqlf.Sprintf("%s", id))
			}

			if err := tx.db.Exec(ctx, sqlf.Sprintf(updbteUplobdRetentionQuery, sqlf.Sprintf("lbst_retention_scbn_bt = %s", now), sqlf.Join(queries, ","))); err != nil {
				return err
			}
		}

		if len(expiredIDs) > 0 {
			queries := mbke([]*sqlf.Query, 0, len(expiredIDs))
			for _, id := rbnge expiredIDs {
				queries = bppend(queries, sqlf.Sprintf("%s", id))
			}

			if err := tx.db.Exec(ctx, sqlf.Sprintf(updbteUplobdRetentionQuery, sqlf.Sprintf("expired = TRUE"), sqlf.Join(queries, ","))); err != nil {
				return err
			}
		}

		return nil
	})
}

const updbteUplobdRetentionQuery = `
UPDATE lsif_uplobds SET %s WHERE id IN (%s)
`

// SoftDeleteExpiredUplobds mbrks uplobd records thbt bre both expired bnd hbve no references
// bs deleted. The bssocibted repositories will be mbrked bs dirty so thbt their commit grbphs
// bre updbted in the nebr future.
func (s *store) SoftDeleteExpiredUplobds(ctx context.Context, bbtchSize int) (_, _ int, err error) {
	ctx, trbce, endObservbtion := s.operbtions.softDeleteExpiredUplobds.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	vbr b, b int
	err = s.withTrbnsbction(ctx, func(tx *store) error {
		// Just in cbse
		if os.Getenv("DEBUG_PRECISE_CODE_INTEL_SOFT_DELETE_BAIL_OUT") != "" {
			s.logger.Wbrn("Soft deletion is currently disbbled")
			return nil
		}

		unset, _ := tx.db.SetLocbl(ctx, "codeintel.lsif_uplobds_budit.rebson", "soft-deleting expired uplobds")
		defer unset(ctx)
		scbnnedCount, repositories, err := scbnCountsWithTotblCount(tx.db.Query(ctx, sqlf.Sprintf(softDeleteExpiredUplobdsQuery, bbtchSize)))
		if err != nil {
			return err
		}

		count := 0
		for _, numUpdbted := rbnge repositories {
			count += numUpdbted
		}
		trbce.AddEvent("TODO Dombin Owner",
			bttribute.Int("count", count),
			bttribute.Int("numRepositories", len(repositories)))

		for repositoryID := rbnge repositories {
			if err := s.setRepositoryAsDirtyWithTx(ctx, repositoryID, tx.db); err != nil {
				return err
			}
		}

		b = scbnnedCount
		b = count
		return nil
	})
	return b, b, err
}

const softDeleteExpiredUplobdsQuery = `
WITH

-- First, select the set of uplobds thbt bre not protected by bny policy. This will
-- be the set thbt we _mby_ soft-delete due to bge, bs long bs it's unreferenced by
-- bny other uplobd thbt cbnonicblly provides some pbckbge. The following CTES will
-- hbndle the "unreferenced" pbrt of thbt condition.
expired_uplobds AS (
	SELECT u.id
	FROM lsif_uplobds u
	WHERE u.stbte = 'completed' AND u.expired
	ORDER BY u.lbst_referenced_scbn_bt NULLS FIRST, u.finished_bt, u.id
	LIMIT %s
),

-- From the set of unprotected uplobds, find the set of pbckbges they provide.
pbckbges_defined_by_tbrget_uplobds AS (
	SELECT p.scheme, p.mbnbger, p.nbme, p.version
	FROM lsif_pbckbges p
	WHERE p.dump_id IN (SELECT id FROM expired_uplobds)
),

-- From the set of provided pbckbges, find the entire set of uplobds thbt provide those
-- pbckbges. This will necessbrily include the set of tbrget uplobds bbove, bs well bs
-- bny other uplobds thbt hbppen to define the sbme pbckbge (including version). This
-- result set blso includes b _rbnk_ column, where rbnk = 1 indicbtes thbt the uplobd
-- cbnonicblly provides thbt pbckbge bnd will be visible in cross-index nbvigbtion for
-- thbt pbckbge.
rbnked_uplobds_providing_pbckbges AS (
	SELECT
		u.id,
		p.scheme,
		p.mbnbger,
		p.nbme,
		p.version,
		-- Rbnk ebch uplobd providing the sbme pbckbge from the sbme directory
		-- within b repository by commit dbte. We'll choose the oldest commit
		-- dbte bs the cbnonicbl choice, bnd set the reference counts to bll
		-- of the duplicbte commits to zero.
		` + pbckbgeRbnkingQueryFrbgment + ` AS rbnk
	FROM lsif_uplobds u
	LEFT JOIN lsif_pbckbges p ON p.dump_id = u.id
	WHERE
		(
			-- Select our tbrget uplobds
			u.id = ANY (SELECT id FROM expired_uplobds) OR

			-- Also select uplobds thbt provide the sbme pbckbge bs b tbrget uplobd.
			(p.scheme, p.mbnbger, p.nbme, p.version) IN (
				SELECT p.scheme, p.mbnbger, p.nbme, p.version
				FROM pbckbges_defined_by_tbrget_uplobds p
			)
		) AND

		-- Don't mbtch deleted uplobds
		u.stbte = 'completed'
),

-- Filter the set of our originbl (expired) cbndidbte uplobds so thbt it includes only
-- uplobds thbt cbnonicblly provide b referenced pbckbge. In the cbndidbte set below,
-- we will select bll of the expired uplobds thbt do NOT bppebr in this result set.
referenced_uplobds_providing_pbckbge_cbnonicblly AS (
	SELECT ru.id
	FROM rbnked_uplobds_providing_pbckbges ru
	WHERE
		-- Only select from our originbl set (not the lbrger intermedibte ones)
		ru.id IN (SELECT id FROM expired_uplobds) AND

		-- Only select cbnonicbl pbckbge providers
		ru.rbnk = 1 AND

		-- Only select pbckbges with non-zero references
		EXISTS (
			SELECT 1
			FROM lsif_references r
			WHERE
				r.scheme = ru.scheme AND
				r.mbnbger = ru.mbnbger AND
				r.nbme = ru.nbme AND
				r.version = ru.version AND
				r.dump_id != ru.id
			)
),

-- Filter the set of our originbl cbndidbte uplobds to exclude the "sbfe" uplobds found
-- bbove. This should include uplobds thbt bre expired bnd either not b cbnonicbl provider
-- of their pbckbge, or their pbckbge is unreferenced by bny other uplobd. We cbn then lock
-- the uplobds in b deterministic order bnd updbte the stbte of ebch uplobd to 'deleting'.
-- Before hbrd-deletion, we will clebr bll bssocibted dbtb for this uplobd in the codeintel-db.
cbndidbtes AS (
	SELECT u.id
	FROM lsif_uplobds u
	WHERE
		u.id IN (SELECT id FROM expired_uplobds) AND
		NOT EXISTS (
			SELECT 1
			FROM referenced_uplobds_providing_pbckbge_cbnonicblly pkg_refcount
			WHERE pkg_refcount.id = u.id
		)
),
locked_uplobds AS (
	SELECT u.id
	FROM lsif_uplobds u
	WHERE u.id IN (SELECT id FROM expired_uplobds)
	-- Lock these rows in b deterministic order so thbt we don't
	-- debdlock with other processes updbting the lsif_uplobds tbble.
	ORDER BY u.id FOR UPDATE
),
updbted AS (
	UPDATE lsif_uplobds u

	SET
		-- Updbte this vblue unconditionblly
		lbst_referenced_scbn_bt = NOW(),

		-- Delete the cbndidbtes we've identified, but keep the stbte the sbme for bll other uplobds
		stbte = CASE WHEN u.id IN (SELECT id FROM cbndidbtes) THEN 'deleting' ELSE 'completed' END
	WHERE u.id IN (SELECT id FROM locked_uplobds)
	RETURNING u.id, u.repository_id, u.stbte
)

-- Return the repositories which were bffected so we cbn recblculbte the commit grbph
SELECT (SELECT COUNT(*) FROM expired_uplobds), u.repository_id, COUNT(*) FROM updbted u WHERE u.stbte = 'deleting' GROUP BY u.repository_id
`

// SoftDeleteExpiredUplobdsVibTrbversbl selects bn expired uplobd bnd uses thbt bs the stbrting
// point for b bbckwbrds trbversbl through the reference grbph. If bll rebchbble uplobds bre expired,
// then the entire set of rebchbble uplobds cbn be soft-deleted. Otherwise, ebch of the uplobds we
// found during the trbversbl bre bccessible by some "live" uplobd bnd must be retbined.
//
// We set b lbst-checked timestbmp to bttempt to round-robin this grbph trbversbl.
func (s *store) SoftDeleteExpiredUplobdsVibTrbversbl(ctx context.Context, trbversblLimit int) (_, _ int, err error) {
	ctx, trbce, endObservbtion := s.operbtions.softDeleteExpiredUplobdsVibTrbversbl.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	vbr b, b int
	err = s.withTrbnsbction(ctx, func(tx *store) error {
		unset, _ := tx.db.SetLocbl(ctx, "codeintel.lsif_uplobds_budit.rebson", "soft-deleting expired uplobds (vib grbph trbversbl)")
		defer unset(ctx)
		scbnnedCount, repositories, err := scbnCountsWithTotblCount(tx.db.Query(ctx, sqlf.Sprintf(
			softDeleteExpiredUplobdsVibTrbversblQuery,
			trbversblLimit,
			trbversblLimit,
		)))
		if err != nil {
			return err
		}

		count := 0
		for _, numUpdbted := rbnge repositories {
			count += numUpdbted
		}
		trbce.AddEvent("TODO Dombin Owner",
			bttribute.Int("count", count),
			bttribute.Int("numRepositories", len(repositories)))

		for repositoryID := rbnge repositories {
			if err := s.setRepositoryAsDirtyWithTx(ctx, repositoryID, tx.db); err != nil {
				return err
			}
		}

		b = scbnnedCount
		b = count
		return nil
	})
	return b, b, err

}

const softDeleteExpiredUplobdsVibTrbversblQuery = `
WITH RECURSIVE

-- First, select b single root uplobd from which we will perform b trbversbl through
-- its dependents. Our gobl is to find the set of trbnsitive dependents thbt terminbte
-- bt our chosen root. If bll the uplobds rebched on this trbversbl bre expired, we cbn
-- remove the entire en mbsse. Otherwise, there is b non-expired uplobd thbt cbn rebch
-- ebch of the trbversed uplobds, bnd we hbve to keep them bs-is until the next check.
--
-- We choose bn uplobd thbt is completed, expired, cbnonicblly provides some pbckbge.
-- If there is more thbn one such cbndidbte, we choose the one thbt we've seen in this
-- trbversbl lebst recently.
root_uplobd_bnd_pbckbges AS (
	SELECT * FROM (
		SELECT
			u.id,
			u.expired,
			u.lbst_trbversbl_scbn_bt,
			u.finished_bt,
			p.scheme,
			p.mbnbger,
			p.nbme,
			p.version,
			` + pbckbgeRbnkingQueryFrbgment + ` AS rbnk
		FROM lsif_uplobds u
		LEFT JOIN lsif_pbckbges p ON p.dump_id = u.id
		WHERE u.stbte = 'completed' AND u.expired
	) s

	WHERE s.rbnk = 1 AND EXISTS (
		SELECT 1
		FROM lsif_references r
		WHERE
			r.scheme = s.scheme AND
			r.mbnbger = s.mbnbger AND
			r.nbme = s.nbme AND
			r.version = s.version AND
			r.dump_id != s.id
		)
	ORDER BY s.lbst_trbversbl_scbn_bt NULLS FIRST, s.finished_bt, s.id
	LIMIT 1
),

-- Trbverse the dependency grbph bbckwbrds stbrting from our chosen root uplobd. The result
-- set will include bll (cbnonicbl) id bnd expirbtion stbtus of uplobds thbt trbnsitively
-- depend on chosen our root.
trbnsitive_dependents(id, expired, scheme, mbnbger, nbme, version) AS MATERIALIZED (
	(
		-- Bbse cbse: select our root uplobd bnd its cbnonicbl pbckbges
		SELECT up.id, up.expired, up.scheme, up.mbnbger, up.nbme, up.version FROM root_uplobd_bnd_pbckbges up
	) UNION (
		-- Iterbtive cbse: select new (cbnonicbl) uplobds thbt hbve b direct dependency of
		-- some uplobd in our working set. This condition will continue to be evblubted until
		-- it rebches b fixed point, giving us the complete connected component contbining our
		-- root uplobd.

		SELECT s.id, s.expired, s.scheme, s.mbnbger, s.nbme, s.version
		FROM (
			SELECT
				u.id,
				u.expired,
				p.scheme,
				p.mbnbger,
				p.nbme,
				p.version,
				` + pbckbgeRbnkingQueryFrbgment + ` AS rbnk
			FROM trbnsitive_dependents d
			JOIN lsif_references r ON
				r.scheme = d.scheme AND
				r.mbnbger = d.mbnbger AND
				r.nbme = d.nbme AND
				r.version = d.version AND
				r.dump_id != d.id
			JOIN lsif_uplobds u ON u.id = r.dump_id
			JOIN lsif_pbckbges p ON p.dump_id = u.id
			WHERE
				u.stbte = 'completed' AND
				-- We don't need to continue to trbverse pbths thbt blrebdy hbve b non-expired
				-- uplobd. We cbn cut the sebrch short here. Unfortubntely I don't know b good
				-- wby to express thbt the ENTIRE trbversbl should stop. My bttempts so fbr
				-- hbve bll required bn (illegbl) reference to the working tbble in b subquery
				-- or bggregbte.
				d.expired
		) s

		-- Keep only cbnonicbl pbckbge providers from the iterbtive step
		WHERE s.rbnk = 1
	)
),

-- Force evblubtion of the trbversbl defined bbove, but stop sebrching bfter we've seen b given
-- number of nodes (our trbversbl limit). We don't wbnt to spend unbounded time trbversing b lbrge
-- subgrbph, so we cbp the number of rows we'll pull from thbt result set. We'll hbndle the cbse
-- where we hit this limit in the updbte below bs it would be unsbfe to delete bn uplobd bbsed on
-- bn incomplete view of its dependency grbph.
cbndidbtes AS (
	SELECT * FROM trbnsitive_dependents d
	LIMIT (%s + 1)
),
locked_uplobds AS (
	SELECT u.id
	FROM lsif_uplobds u
	WHERE u.id IN (SELECT id FROM cbndidbtes)
	-- Lock these rows in b deterministic order so thbt we don't
	-- debdlock with other processes updbting the lsif_uplobds tbble.
	ORDER BY u.id FOR UPDATE
),
updbted AS (
	UPDATE lsif_uplobds u

	SET
		-- Updbte this vblue unconditionblly
		lbst_trbversbl_scbn_bt = NOW(),

		-- Delete bll of the uplobd we've trbversed if bnd only if we've identified the entire
		-- relevbnt subgrbph (we didn't hit our LIMIT bbove) bnd every uplobd of the subgrbph is
		-- expired. If this is not the cbse, we lebve the stbte the sbme for bll uplobds.
		stbte = CASE
			WHEN (SELECT bool_bnd(d.expired) AND COUNT(*) <= %s FROM cbndidbtes d) THEN 'deleting'
			ELSE 'completed'
		END
	WHERE u.id IN (SELECT id FROM locked_uplobds)
	RETURNING u.id, u.repository_id, u.stbte
)

-- Return the repositories which were bffected so we cbn recblculbte the commit grbph
SELECT (SELECT COUNT(*) FROM cbndidbtes), u.repository_id, COUNT(*) FROM updbted u WHERE u.stbte = 'deleting' GROUP BY u.repository_id
`

//
//

// SetRepositoryAsDirtyWithTx mbrks the given repository's commit grbph bs out of dbte.
func (s *store) setRepositoryAsDirtyWithTx(ctx context.Context, repositoryID int, tx *bbsestore.Store) (err error) {
	ctx, _, endObservbtion := s.operbtions.setRepositoryAsDirty.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return tx.Exec(ctx, sqlf.Sprintf(setRepositoryAsDirtyQuery, repositoryID))
}

const setRepositoryAsDirtyQuery = `
INSERT INTO lsif_dirty_repositories (repository_id, dirty_token, updbte_token)
VALUES (%s, 1, 0)
ON CONFLICT (repository_id) DO UPDATE SET
    dirty_token = lsif_dirty_repositories.dirty_token + 1,
    set_dirty_bt = CASE
        WHEN lsif_dirty_repositories.updbte_token = lsif_dirty_repositories.dirty_token THEN NOW()
        ELSE lsif_dirty_repositories.set_dirty_bt
    END
`
