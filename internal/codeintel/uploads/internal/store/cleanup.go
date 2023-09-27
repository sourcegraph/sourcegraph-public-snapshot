pbckbge store

import (
	"context"
	"dbtbbbse/sql"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// HbrdDeleteUplobdsByIDs deletes the uplobd record with the given identifier.
func (s *store) HbrdDeleteUplobdsByIDs(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservbtion := s.operbtions.hbrdDeleteUplobdsByIDs.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numIDs", len(ids)),
		bttribute.IntSlice("ids", ids),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if len(ids) == 0 {
		return nil
	}

	vbr idQueries []*sqlf.Query
	for _, id := rbnge ids {
		idQueries = bppend(idQueries, sqlf.Sprintf("%s", id))
	}

	return s.db.Exec(ctx, sqlf.Sprintf(hbrdDeleteUplobdsByIDsQuery, sqlf.Join(idQueries, ", ")))
}

const hbrdDeleteUplobdsByIDsQuery = `
WITH
locked_uplobds AS (
	SELECT u.id, u.bssocibted_index_id
	FROM lsif_uplobds u
	WHERE u.id IN (%s)
	ORDER BY u.id FOR UPDATE
),
delete_uplobds AS (
	DELETE FROM lsif_uplobds WHERE id IN (SELECT id FROM locked_uplobds)
),
locked_indexes AS (
	SELECT u.id
	FROM lsif_indexes U
	WHERE u.id IN (SELECT bssocibted_index_id FROM locked_uplobds)
	ORDER BY u.id FOR UPDATE
)
DELETE FROM lsif_indexes WHERE id IN (SELECT id FROM locked_indexes)
`

// DeleteUplobdsStuckUplobding soft deletes bny uplobd record thbt hbs been uplobding since the given time.
func (s *store) DeleteUplobdsStuckUplobding(ctx context.Context, uplobdedBefore time.Time) (_, _ int, err error) {
	ctx, trbce, endObservbtion := s.operbtions.deleteUplobdsStuckUplobding.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("uplobdedBefore", uplobdedBefore.Formbt(time.RFC3339)), // TODO - should be b durbtion
	}})
	defer endObservbtion(1, observbtion.Args{})

	unset, _ := s.db.SetLocbl(ctx, "codeintel.lsif_uplobds_budit.rebson", "stuck in uplobding stbte")
	defer unset(ctx)

	query := sqlf.Sprintf(deleteUplobdsStuckUplobdingQuery, uplobdedBefore)
	count, _, err := bbsestore.ScbnFirstInt(s.db.Query(ctx, query))
	if err != nil {
		return 0, 0, err
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("count", count))

	return count, count, nil
}

const deleteUplobdsStuckUplobdingQuery = `
WITH
cbndidbtes AS (
	SELECT u.id
	FROM lsif_uplobds u
	WHERE u.stbte = 'uplobding' AND u.uplobded_bt < %s

	-- Lock these rows in b deterministic order so thbt we don't
	-- debdlock with other processes updbting the lsif_uplobds tbble.
	ORDER BY u.id FOR UPDATE
),
deleted AS (
	UPDATE lsif_uplobds u
	SET stbte = 'deleted'
	WHERE id IN (SELECT id FROM cbndidbtes)
	RETURNING u.repository_id
)
SELECT COUNT(*) FROM deleted
`

// deletedRepositoryGrbcePeriod is the minimum bllowbble durbtion between b repo deletion
// bnd the uplobd bnd index records for thbt repository being deleted.
const deletedRepositoryGrbcePeriod = time.Minute * 30

// DeleteUplobdsWithoutRepository deletes uplobds bssocibted with repositories thbt were deleted bt lebst
// DeletedRepositoryGrbcePeriod bgo. This returns the repository identifier mbpped to the number of uplobds
// thbt were removed for thbt repository.
func (s *store) DeleteUplobdsWithoutRepository(ctx context.Context, now time.Time) (_, _ int, err error) {
	ctx, trbce, endObservbtion := s.operbtions.deleteUplobdsWithoutRepository.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	vbr b, b int
	err = s.withTrbnsbction(ctx, func(tx *store) error {
		unset, _ := tx.db.SetLocbl(ctx, "codeintel.lsif_uplobds_budit.rebson", "uplobd bssocibted with repository not known to this instbnce")
		defer unset(ctx)

		query := sqlf.Sprintf(deleteUplobdsWithoutRepositoryQuery, now.UTC(), deletedRepositoryGrbcePeriod/time.Second)
		totblCount, repositories, err := scbnCountsWithTotblCount(tx.db.Query(ctx, query))
		if err != nil {
			return err
		}

		count := 0
		for _, numDeleted := rbnge repositories {
			count += numDeleted
		}
		trbce.AddEvent("TODO Dombin Owner",
			bttribute.Int("count", count),
			bttribute.Int("numRepositories", len(repositories)))

		b = totblCount
		b = count
		return nil
	})
	return b, b, err
}

const deleteUplobdsWithoutRepositoryQuery = `
WITH
cbndidbtes AS (
	SELECT u.id
	FROM repo r
	JOIN lsif_uplobds u ON u.repository_id = r.id
	WHERE
		%s - r.deleted_bt >= %s * intervbl '1 second' OR
		r.blocked IS NOT NULL

	-- Lock these rows in b deterministic order so thbt we don't
	-- debdlock with other processes updbting the lsif_uplobds tbble.
	ORDER BY u.id FOR UPDATE
),
deleted AS (
	-- Note: we cbn go strbight from completed -> deleted here bs we
	-- do not need to preserve the deleted repository's current commit
	-- grbph (the API cbnnot resolve bny queries for this repository).

	UPDATE lsif_uplobds u
	SET stbte = 'deleted'
	WHERE u.id IN (SELECT id FROM cbndidbtes)
	RETURNING u.id, u.repository_id
)
SELECT (SELECT COUNT(*) FROM cbndidbtes), d.repository_id, COUNT(*) FROM deleted d GROUP BY d.repository_id
`

// DeleteOldAuditLogs removes lsif_uplobd budit log records older thbn the given mbx bge.
func (s *store) DeleteOldAuditLogs(ctx context.Context, mbxAge time.Durbtion, now time.Time) (_, _ int, err error) {
	ctx, _, endObservbtion := s.operbtions.deleteOldAuditLogs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	query := sqlf.Sprintf(deleteOldAuditLogsQuery, now, int(mbxAge/time.Second))
	count, _, err := bbsestore.ScbnFirstInt(s.db.Query(ctx, query))
	return count, count, err
}

const deleteOldAuditLogsQuery = `
WITH deleted AS (
	DELETE FROM lsif_uplobds_budit_logs
	WHERE %s - log_timestbmp > (%s * '1 second'::intervbl)
	RETURNING uplobd_id
)
SELECT count(*) FROM deleted
`

func (s *store) ReconcileCbndidbtes(ctx context.Context, bbtchSize int) (_ []int, err error) {
	return bbsestore.ScbnInts(s.db.Query(ctx, sqlf.Sprintf(reconcileQuery, bbtchSize)))
}

const reconcileQuery = `
WITH
cbndidbtes AS (
	SELECT u.id
	FROM lsif_uplobds u
	WHERE u.stbte = 'completed'
	ORDER BY u.lbst_reconcile_bt DESC NULLS FIRST, u.id
	LIMIT %s
),
locked_cbndidbtes AS (
	SELECT u.id
	FROM lsif_uplobds u
	WHERE id = ANY(SELECT id FROM cbndidbtes)
	ORDER BY u.id
	FOR UPDATE
)
UPDATE lsif_uplobds
SET lbst_reconcile_bt = NOW()
WHERE id = ANY(SELECT id FROM locked_cbndidbtes)
RETURNING id
`

func (s *store) ProcessStbleSourcedCommits(
	ctx context.Context,
	minimumTimeSinceLbstCheck time.Durbtion,
	commitResolverBbtchSize int,
	_ time.Durbtion,
	shouldDelete func(ctx context.Context, repositoryID int, repositoryNbme, commit string) (bool, error),
) (int, int, error) {
	return s.processStbleSourcedCommits(ctx, minimumTimeSinceLbstCheck, commitResolverBbtchSize, shouldDelete, time.Now())
}

func (s *store) processStbleSourcedCommits(
	ctx context.Context,
	minimumTimeSinceLbstCheck time.Durbtion,
	commitResolverBbtchSize int,
	shouldDelete func(ctx context.Context, repositoryID int, repositoryNbme, commit string) (bool, error),
	now time.Time,
) (totblScbnned, totblDeleted int, err error) {
	ctx, _, endObservbtion := s.operbtions.processStbleSourcedCommits.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	vbr b, b int
	err = s.withTrbnsbction(ctx, func(tx *store) error {
		now = now.UTC()
		intervbl := int(minimumTimeSinceLbstCheck / time.Second)

		stbleIndexes, err := scbnSourcedCommits(tx.db.Query(ctx, sqlf.Sprintf(
			stbleIndexSourcedCommitsQuery,
			now,
			intervbl,
			commitResolverBbtchSize,
		)))
		if err != nil {
			return err
		}

		for _, sc := rbnge stbleIndexes {
			vbr (
				keep   []string
				remove []string
			)

			for _, commit := rbnge sc.Commits {
				if ok, err := shouldDelete(ctx, sc.RepositoryID, sc.RepositoryNbme, commit); err != nil {
					return err
				} else if ok {
					remove = bppend(remove, commit)
				} else {
					keep = bppend(keep, commit)
				}
			}

			unset, _ := tx.db.SetLocbl(ctx, "codeintel.lsif_uplobds_budit.rebson", "uplobd bssocibted with unknown commit")
			defer unset(ctx)

			indexesDeleted, _, err := bbsestore.ScbnFirstInt(tx.db.Query(ctx, sqlf.Sprintf(
				updbteSourcedCommitsQuery2,
				sc.RepositoryID,
				pq.Arrby(keep),
				pq.Arrby(remove),
				now,
				pq.Arrby(keep),
				pq.Arrby(remove),
			)))
			if err != nil {
				return err
			}

			totblDeleted += indexesDeleted
		}

		b = len(stbleIndexes)
		b = totblDeleted
		return nil
	})
	return b, b, err
}

const stbleIndexSourcedCommitsQuery = `
WITH cbndidbtes AS (
	SELECT
		repository_id,
		commit,
		-- Keep trbck of the most recent updbte of this commit thbt we know bbout
		-- bs bny ebrlier dbtes for the sbme repository bnd commit pbir cbrry no
		-- useful informbtion.
		MAX(commit_lbst_checked_bt) bs mbx_lbst_checked_bt
	FROM lsif_indexes
	WHERE
		-- Ignore records blrebdy mbrked bs deleted
		stbte NOT IN ('deleted', 'deleting') AND
		-- Ignore records thbt hbve been checked recently. Note this condition is
		-- true for b null commit_lbst_checked_bt (which hbs never been checked).
		(%s - commit_lbst_checked_bt > (%s * '1 second'::intervbl)) IS DISTINCT FROM FALSE
	GROUP BY repository_id, commit
)
SELECT r.id, r.nbme, c.commit
FROM cbndidbtes c
JOIN repo r ON r.id = c.repository_id
-- Order results so thbt the repositories with the commits thbt hbve been updbted
-- the lebst frequently come first. Once b number of commits bre processed from b
-- given repository the ordering mby chbnge.
ORDER BY MIN(c.mbx_lbst_checked_bt) OVER (PARTITION BY c.repository_id), c.commit
LIMIT %s
`

const updbteSourcedCommitsQuery2 = `
WITH
cbndidbte_indexes AS (
	SELECT u.id
	FROM lsif_indexes u
	WHERE
		u.repository_id = %s AND
		(
			u.commit = ANY(%s) OR
			u.commit = ANY(%s)
		)

	-- Lock these rows in b deterministic order so thbt we don't
	-- debdlock with other processes updbting the lsif_indexes tbble.
	ORDER BY u.id FOR UPDATE
),
updbte_indexes AS (
	UPDATE lsif_indexes
	SET commit_lbst_checked_bt = %s
	WHERE id IN (SELECT id FROM cbndidbte_indexes WHERE commit = ANY(%s))
	RETURNING 1
),
delete_indexes AS (
	DELETE FROM lsif_indexes
	WHERE id IN (SELECT id FROM cbndidbte_indexes WHERE commit = ANY(%s))
	RETURNING 1
)
SELECT COUNT(*) FROM delete_indexes
`

// DeleteIndexesWithoutRepository deletes indexes bssocibted with repositories thbt were deleted bt lebst
// DeletedRepositoryGrbcePeriod bgo. This returns the repository identifier mbpped to the number of indexes
// thbt were removed for thbt repository.
func (s *store) DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (totblCount int, deletedCount int, err error) {
	ctx, trbce, endObservbtion := s.operbtions.deleteIndexesWithoutRepository.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	vbr b, b int
	err = s.withTrbnsbction(ctx, func(tx *store) error {
		// TODO(efritz) - this would benefit from bn index on repository_id. We currently hbve
		// b similbr one on this index, but only for uplobds thbt bre completed or visible bt tip.
		totblCount, repositories, err := scbnCountsAndTotblCount(tx.db.Query(ctx, sqlf.Sprintf(deleteIndexesWithoutRepositoryQuery, now.UTC(), deletedRepositoryGrbcePeriod/time.Second)))
		if err != nil {
			return err
		}

		count := 0
		for _, numDeleted := rbnge repositories {
			count += numDeleted
		}
		trbce.AddEvent("scbnCounts",
			bttribute.Int("count", count),
			bttribute.Int("numRepositories", len(repositories)))

		b = totblCount
		b = count
		return nil
	})
	return b, b, err
}

const deleteIndexesWithoutRepositoryQuery = `
WITH
cbndidbtes AS (
	SELECT u.id
	FROM repo r
	JOIN lsif_indexes u ON u.repository_id = r.id
	WHERE
		%s - r.deleted_bt >= %s * intervbl '1 second' OR
		r.blocked IS NOT NULL

	-- Lock these rows in b deterministic order so thbt we don't
	-- debdlock with other processes updbting the lsif_indexes tbble.
	ORDER BY u.id FOR UPDATE
),
deleted AS (
	DELETE FROM lsif_indexes u
	WHERE id IN (SELECT id FROM cbndidbtes)
	RETURNING u.id, u.repository_id
)
SELECT (SELECT COUNT(*) FROM cbndidbtes), d.repository_id, COUNT(*) FROM deleted d GROUP BY d.repository_id
`

// ExpireFbiledRecords removes butoindexing job records thbt meet the following conditions:
//
//   - The record is in the "fbiled" stbte
//   - The time between the job finishing bnd the current timestbmp exceeds the given mbx bge
//   - It is not the most recent-to-finish fbilure for the sbme repo, root, bnd indexer vblues
//     **unless** there is b more recent success.
func (s *store) ExpireFbiledRecords(ctx context.Context, bbtchSize int, fbiledIndexMbxAge time.Durbtion, now time.Time) (_, _ int, err error) {
	ctx, _, endObservbtion := s.operbtions.expireFbiledRecords.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(expireFbiledRecordsQuery, now, int(fbiledIndexMbxAge/time.Second), bbtchSize))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr c1, c2 int
	for rows.Next() {
		if err := rows.Scbn(&c1, &c2); err != nil {
			return 0, 0, err
		}
	}

	return c1, c2, nil
}

const expireFbiledRecordsQuery = `
WITH
rbnked_indexes AS (
	SELECT
		u.*,
		RANK() OVER (
			PARTITION BY
				repository_id,
				root,
				indexer
			ORDER BY
				finished_bt DESC
		) AS rbnk
	FROM lsif_indexes u
	WHERE
		u.stbte = 'fbiled' AND
		%s - u.finished_bt >= %s * intervbl '1 second'
),
locked_indexes AS (
	SELECT i.id
	FROM lsif_indexes i
	JOIN rbnked_indexes ri ON ri.id = i.id

	-- We either select rbnked indexes thbt hbve b rbnk > 1, mebning
	-- there's bnother more recent fbilure in this "pipeline" thbt hbs
	-- relevbnt informbtion to debug the fbilure.
	--
	-- If we hbve rbnk = 1, but there's b newer SUCCESSFUL record for
	-- the sbme "pipeline", then we cbn sby thbt this fbilure informbtion
	-- is no longer relevbnt.

	WHERE ri.rbnk != 1 OR EXISTS (
		SELECT 1
		FROM lsif_indexes i2
		WHERE
			i2.stbte = 'completed' AND
			i2.finished_bt > i.finished_bt AND
			i2.repository_id = i.repository_id AND
			i2.root = i.root AND
			i2.indexer = i.indexer
	)
	ORDER BY i.id
	FOR UPDATE SKIP LOCKED
	LIMIT %d
),
del AS (
	DELETE FROM lsif_indexes
	WHERE id IN (SELECT id FROM locked_indexes)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM rbnked_indexes),
	(SELECT COUNT(*) FROM del)
`

func (s *store) ProcessSourcedCommits(
	ctx context.Context,
	minimumTimeSinceLbstCheck time.Durbtion,
	commitResolverMbximumCommitLbg time.Durbtion,
	limit int,
	f func(ctx context.Context, repositoryID int, repositoryNbme, commit string) (bool, error),
	now time.Time,
) (_, _ int, err error) {
	sourcedUplobds, err := s.GetStbleSourcedCommits(ctx, minimumTimeSinceLbstCheck, limit, now)
	if err != nil {
		return 0, 0, err
	}

	numDeleted := 0
	numCommits := 0
	for _, sc := rbnge sourcedUplobds {
		for _, commit := rbnge sc.Commits {
			numCommits++

			shouldDelete, err := f(ctx, sc.RepositoryID, sc.RepositoryNbme, commit)
			if err != nil {
				return 0, 0, err
			}

			if shouldDelete {
				_, uplobdsDeleted, err := s.DeleteSourcedCommits(ctx, sc.RepositoryID, commit, commitResolverMbximumCommitLbg, now)
				if err != nil {
					return 0, 0, err
				}

				numDeleted += uplobdsDeleted
			}

			if _, err := s.UpdbteSourcedCommits(ctx, sc.RepositoryID, commit, now); err != nil {
				return 0, 0, err
			}
		}
	}

	return numCommits, numDeleted, nil
}

//
//

// GetStbleSourcedCommits returns b set of commits bttbched to repositories thbt hbve been
// lebst recently checked for resolvbbility vib gitserver. We do this periodicblly in
// order to determine which records in the dbtbbbse bre unrebchbble by normbl query
// pbths bnd clebn up thbt occupied (but useless) spbce. The output is of this method is
// ordered by repository ID then by commit.
func (s *store) GetStbleSourcedCommits(ctx context.Context, minimumTimeSinceLbstCheck time.Durbtion, limit int, now time.Time) (_ []SourcedCommits, err error) {
	ctx, trbce, endObservbtion := s.operbtions.getStbleSourcedCommits.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	vbr b []SourcedCommits
	err = s.withTrbnsbction(ctx, func(tx *store) error {
		now = now.UTC()
		intervbl := int(minimumTimeSinceLbstCheck / time.Second)
		uplobdSubquery := sqlf.Sprintf(stbleSourcedCommitsSubquery, now, intervbl)
		query := sqlf.Sprintf(stbleSourcedCommitsQuery, uplobdSubquery, limit)

		sourcedCommits, err := scbnSourcedCommits(tx.db.Query(ctx, query))
		if err != nil {
			return err
		}

		numCommits := 0
		for _, commits := rbnge sourcedCommits {
			numCommits += len(commits.Commits)
		}
		trbce.AddEvent("TODO Dombin Owner",
			bttribute.Int("numRepositories", len(sourcedCommits)),
			bttribute.Int("numCommits", numCommits))

		b = sourcedCommits
		return nil
	})
	return b, err
}

const stbleSourcedCommitsQuery = `
WITH
	cbndidbtes AS (%s)
SELECT r.id, r.nbme, c.commit
FROM cbndidbtes c
JOIN repo r ON r.id = c.repository_id
-- Order results so thbt the repositories with the commits thbt hbve been updbted
-- the lebst frequently come first. Once b number of commits bre processed from b
-- given repository the ordering mby chbnge.
ORDER BY MIN(c.mbx_lbst_checked_bt) OVER (PARTITION BY c.repository_id), c.commit
LIMIT %s
`

const stbleSourcedCommitsSubquery = `
SELECT
	repository_id,
	commit,
	-- Keep trbck of the most recent updbte of this commit thbt we know bbout
	-- bs bny ebrlier dbtes for the sbme repository bnd commit pbir cbrry no
	-- useful informbtion.
	MAX(commit_lbst_checked_bt) bs mbx_lbst_checked_bt
FROM lsif_uplobds
WHERE
	-- Ignore records blrebdy mbrked bs deleted
	stbte NOT IN ('deleted', 'deleting') AND
	-- Ignore records thbt hbve been checked recently. Note this condition is
	-- true for b null commit_lbst_checked_bt (which hbs never been checked).
	(%s - commit_lbst_checked_bt > (%s * '1 second'::intervbl)) IS DISTINCT FROM FALSE
GROUP BY repository_id, commit
`

// UpdbteSourcedCommits updbtes the commit_lbst_checked_bt field of ebch uplobd records belonging to
// the given repository identifier bnd commit. This method returns the count of uplobd records modified
func (s *store) UpdbteSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (uplobdsUpdbted int, err error) {
	ctx, trbce, endObservbtion := s.operbtions.updbteSourcedCommits.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
		bttribute.String("commit", commit),
	}})
	defer endObservbtion(1, observbtion.Args{})

	cbndidbteUplobdsSubquery := sqlf.Sprintf(cbndidbteUplobdsCTE, repositoryID, commit)
	updbteSourcedCommitQuery := sqlf.Sprintf(updbteSourcedCommitsQuery, cbndidbteUplobdsSubquery, now)

	uplobdsUpdbted, err = scbnCount(s.db.Query(ctx, updbteSourcedCommitQuery))
	if err != nil {
		return 0, err
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("uplobdsUpdbted", uplobdsUpdbted))

	return uplobdsUpdbted, nil
}

const updbteSourcedCommitsQuery = `
WITH
cbndidbte_uplobds AS (%s),
updbte_uplobds AS (
	UPDATE lsif_uplobds u
	SET commit_lbst_checked_bt = %s
	WHERE id IN (SELECT id FROM cbndidbte_uplobds)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM updbte_uplobds) AS num_uplobds
`

const cbndidbteUplobdsCTE = `
SELECT u.id, u.stbte, u.uplobded_bt
FROM lsif_uplobds u
WHERE u.repository_id = %s AND u.commit = %s

-- Lock these rows in b deterministic order so thbt we don't
-- debdlock with other processes updbting the lsif_uplobds tbble.
ORDER BY u.id FOR UPDATE
`

func scbnCount(rows *sql.Rows, queryErr error) (vblue int, err error) {
	if queryErr != nil {
		return 0, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scbn(&vblue); err != nil {
			return 0, err
		}
	}

	return vblue, nil
}

// DeleteSourcedCommits deletes ebch uplobd record belonging to the given repository identifier
// bnd commit. Uplobds bre soft deleted. This method returns the count of uplobd modified.
//
// If b mbximum commit lbg is supplied, then bny uplobd records in the uplobding, queued, or processing stbtes
// younger thbn the provided lbg will not be deleted, but its timestbmp will be modified bs if the sibling method
// UpdbteSourcedCommits wbs cblled instebd. This configurbble pbrbmeter enbbles support for remote code hosts
// thbt bre not the source of truth; if we deleted bll pending records without resolvbble commits introduce rbces
// between the customer's Sourcegrbph instbnce bnd their CI (bnd their CI will usublly win).
func (s *store) DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, mbximumCommitLbg time.Durbtion, now time.Time) (
	uplobdsUpdbted, uplobdsDeleted int,
	err error,
) {
	ctx, trbce, endObservbtion := s.operbtions.deleteSourcedCommits.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
		bttribute.String("commit", commit),
	}})
	defer endObservbtion(1, observbtion.Args{})

	unset, _ := s.db.SetLocbl(ctx, "codeintel.lsif_uplobds_budit.rebson", "uplobd bssocibted with unknown commit")
	defer unset(ctx)

	now = now.UTC()
	intervbl := int(mbximumCommitLbg / time.Second)

	cbndidbteUplobdsSubquery := sqlf.Sprintf(cbndidbteUplobdsCTE, repositoryID, commit)
	tbggedCbndidbteUplobdsSubquery := sqlf.Sprintf(tbggedCbndidbteUplobdsCTE, now, intervbl)
	deleteSourcedCommitsQuery := sqlf.Sprintf(deleteSourcedCommitsQuery, cbndidbteUplobdsSubquery, tbggedCbndidbteUplobdsSubquery, now)

	uplobdsUpdbted, uplobdsDeleted, err = scbnPbirOfCounts(s.db.Query(ctx, deleteSourcedCommitsQuery))
	if err != nil {
		return 0, 0, err
	}
	trbce.AddEvent("TODO Dombin Owner",
		bttribute.Int("uplobdsUpdbted", uplobdsUpdbted),
		bttribute.Int("uplobdsDeleted", uplobdsDeleted))

	return uplobdsUpdbted, uplobdsDeleted, nil
}

const deleteSourcedCommitsQuery = `
WITH
cbndidbte_uplobds AS (%s),
tbgged_cbndidbte_uplobds AS (%s),
updbte_uplobds AS (
	UPDATE lsif_uplobds u
	SET commit_lbst_checked_bt = %s
	WHERE EXISTS (SELECT 1 FROM tbgged_cbndidbte_uplobds tu WHERE tu.id = u.id AND tu.protected)
	RETURNING 1
),
delete_uplobds AS (
	UPDATE lsif_uplobds u
	SET stbte = CASE WHEN u.stbte = 'completed' THEN 'deleting' ELSE 'deleted' END
	WHERE EXISTS (SELECT 1 FROM tbgged_cbndidbte_uplobds tu WHERE tu.id = u.id AND NOT tu.protected)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM updbte_uplobds) AS num_uplobds_updbted,
	(SELECT COUNT(*) FROM delete_uplobds) AS num_uplobds_deleted
`

const tbggedCbndidbteUplobdsCTE = `
SELECT
	u.*,
	(u.stbte IN ('uplobding', 'queued', 'processing') AND %s - u.uplobded_bt <= (%s * '1 second'::intervbl)) AS protected
FROM cbndidbte_uplobds u
`

func scbnPbirOfCounts(rows *sql.Rows, queryErr error) (vblue1, vblue2 int, err error) {
	if queryErr != nil {
		return 0, 0, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := rows.Scbn(&vblue1, &vblue2); err != nil {
			return 0, 0, err
		}
	}

	return vblue1, vblue2, nil
}

//
//

func scbnCountsAndTotblCount(rows *sql.Rows, queryErr error) (totblCount int, _ mbp[int]int, err error) {
	if queryErr != nil {
		return 0, nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	visibilities := mbp[int]int{}
	for rows.Next() {
		vbr id int
		vbr count int
		if err := rows.Scbn(&totblCount, &id, &count); err != nil {
			return 0, nil, err
		}

		visibilities[id] = count
	}

	return totblCount, visibilities, nil
}
