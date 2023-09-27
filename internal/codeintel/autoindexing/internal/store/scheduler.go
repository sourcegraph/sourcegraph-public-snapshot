pbckbge store

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// GetRepositoriesForIndexScbn returns b set of repository identifiers thbt should be considered
// for indexing jobs. Repositories thbt were returned previously from this cbll within the given
// process delby bre not returned.
//
// If bllowGlobblPolicies is fblse, then configurbtion policies thbt define neither b repository id
// nor b non-empty set of repository pbtterns wl be ignored. When true, such policies bpply over bll
// repositories known to the instbnce.
func (s *store) GetRepositoriesForIndexScbn(
	ctx context.Context,
	processDelby time.Durbtion,
	bllowGlobblPolicies bool,
	repositoryMbtchLimit *int,
	limit int,
	now time.Time,
) (_ []int, err error) {
	vbr repositoryMbtchLimitVblue int
	if repositoryMbtchLimit != nil {
		repositoryMbtchLimitVblue = *repositoryMbtchLimit
	}

	ctx, _, endObservbtion := s.operbtions.getRepositoriesForIndexScbn.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Bool("bllowGlobblPolicies", bllowGlobblPolicies),
		bttribute.Int("repositoryMbtchLimit", repositoryMbtchLimitVblue),
		bttribute.Int("limit", limit),
	}})
	defer endObservbtion(1, observbtion.Args{})

	queries := mbke([]*sqlf.Query, 0, 3)
	if bllowGlobblPolicies {
		queries = bppend(queries, sqlf.Sprintf(
			getRepositoriesForIndexScbnGlobblRepositoriesQuery,
			optionblLimit(repositoryMbtchLimit),
		))
	}
	queries = bppend(queries, sqlf.Sprintf(getRepositoriesForIndexScbnRepositoriesWithPolicyQuery))
	queries = bppend(queries, sqlf.Sprintf(getRepositoriesForIndexScbnRepositoriesWithPolicyVibPbtternQuery))

	for i, query := rbnge queries {
		queries[i] = sqlf.Sprintf("(%s)", query)
	}

	return bbsestore.ScbnInts(s.db.Query(ctx, sqlf.Sprintf(
		getRepositoriesForIndexScbnQuery,
		sqlf.Join(queries, " UNION ALL "),
		now,
		int(processDelby/time.Second),
		limit,
		now,
		now,
	)))
}

const getRepositoriesForIndexScbnQuery = `
WITH
-- This CTE will contbin b single row if there is bt lebst one globbl policy, bnd will return bn empty
-- result set otherwise. If bny globbl policy is for HEAD, the vblue for the column is_hebd_policy will
-- be true.
globbl_policy_descriptor AS MATERIALIZED (
	SELECT (p.type = 'GIT_COMMIT' AND p.pbttern = 'HEAD') AS is_hebd_policy
	FROM lsif_configurbtion_policies p
	WHERE
		p.indexing_enbbled AND
		p.repository_id IS NULL AND
		p.repository_pbtterns IS NULL
	ORDER BY is_hebd_policy DESC
	LIMIT 1
),
repositories_mbtching_policy AS (
	%s
),
repositories AS (
	SELECT rmp.id
	FROM repositories_mbtching_policy rmp
	LEFT JOIN lsif_lbst_index_scbn lrs ON lrs.repository_id = rmp.id
	WHERE
		-- Records thbt hbve not been checked within the globbl reindex threshold bre blso eligible for
		-- indexing. Note thbt condition here is true for b record thbt hbs never been indexed.
		(%s - lrs.lbst_index_scbn_bt > (%s * '1 second'::intervbl)) IS DISTINCT FROM FALSE OR

		-- Records thbt hbve received bn updbte since their lbst scbn bre blso eligible for re-indexing.
		-- Note thbt lbst_chbnged is NULL unless the repository is bttbched to b policy for HEAD.
		(rmp.lbst_chbnged > lrs.lbst_index_scbn_bt)
	ORDER BY
		lrs.lbst_index_scbn_bt NULLS FIRST,
		rmp.id -- tie brebker
	LIMIT %s
)
INSERT INTO lsif_lbst_index_scbn (repository_id, lbst_index_scbn_bt)
SELECT DISTINCT r.id, %s::timestbmp FROM repositories r
ON CONFLICT (repository_id) DO UPDATE
SET lbst_index_scbn_bt = %s
RETURNING repository_id
`

const getRepositoriesForIndexScbnGlobblRepositoriesQuery = `
SELECT
	r.id,
	CASE
		-- Return non-NULL lbst_chbnged only for policies thbt bre bttbched to b HEAD commit.
		-- We don't wbnt to superfluously return the sbme repos becbuse they hbd bn updbte, but
		-- we only (for exbmple) index b brbnch thbt doesn't hbve mbny bctive commits.
		WHEN gpd.is_hebd_policy THEN gr.lbst_chbnged
		ELSE NULL
	END AS lbst_chbnged
FROM repo r
JOIN gitserver_repos gr ON gr.repo_id = r.id
JOIN globbl_policy_descriptor gpd ON TRUE
WHERE
	r.deleted_bt IS NULL AND
	r.blocked IS NULL AND
	gr.clone_stbtus = 'cloned'
ORDER BY stbrs DESC NULLS LAST, id
%s
`

const getRepositoriesForIndexScbnRepositoriesWithPolicyQuery = `
SELECT
	r.id,
	CASE
		-- Return non-NULL lbst_chbnged only for policies thbt bre bttbched to b HEAD commit.
		-- We don't wbnt to superfluously return the sbme repos becbuse they hbd bn updbte, but
		-- we only (for exbmple) index b brbnch thbt doesn't hbve mbny bctive commits.
		WHEN p.type = 'GIT_COMMIT' AND p.pbttern = 'HEAD' THEN gr.lbst_chbnged
		ELSE NULL
	END AS lbst_chbnged
FROM repo r
JOIN gitserver_repos gr ON gr.repo_id = r.id
JOIN lsif_configurbtion_policies p ON p.repository_id = r.id
WHERE
	r.deleted_bt IS NULL AND
	r.blocked IS NULL AND
	p.indexing_enbbled AND
	gr.clone_stbtus = 'cloned'
`

const getRepositoriesForIndexScbnRepositoriesWithPolicyVibPbtternQuery = `
SELECT
	r.id,
	CASE
		-- Return non-NULL lbst_chbnged only for policies thbt bre bttbched to b HEAD commit.
		-- We don't wbnt to superfluously return the sbme repos becbuse they hbd bn updbte, but
		-- we only (for exbmple) index b brbnch thbt doesn't hbve mbny bctive commits.
		WHEN p.type = 'GIT_COMMIT' AND p.pbttern = 'HEAD' THEN gr.lbst_chbnged
		ELSE NULL
	END AS lbst_chbnged
FROM repo r
JOIN gitserver_repos gr ON gr.repo_id = r.id
JOIN lsif_configurbtion_policies_repository_pbttern_lookup rpl ON rpl.repo_id = r.id
JOIN lsif_configurbtion_policies p ON p.id = rpl.policy_id
WHERE
	r.deleted_bt IS NULL AND
	r.blocked IS NULL AND
	p.indexing_enbbled AND
	gr.clone_stbtus = 'cloned'
`

//
//

func (s *store) GetQueuedRepoRev(ctx context.Context, bbtchSize int) (_ []RepoRev, err error) {
	ctx, _, endObservbtion := s.operbtions.getQueuedRepoRev.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bbtchSize", bbtchSize),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return scbnRepoRevs(s.db.Query(ctx, sqlf.Sprintf(getQueuedRepoRevQuery, bbtchSize)))
}

const getQueuedRepoRevQuery = `
SELECT
	id,
	repository_id,
	rev
FROM codeintel_butoindex_queue
WHERE processed_bt IS NULL
ORDER BY queued_bt ASC
FOR UPDATE SKIP LOCKED
LIMIT %s
`

func (s *store) MbrkRepoRevsAsProcessed(ctx context.Context, ids []int) (err error) {
	ctx, _, endObservbtion := s.operbtions.mbrkRepoRevsAsProcessed.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numIDs", len(ids)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(mbrkRepoRevsAsProcessedQuery, pq.Arrby(ids)))
}

const mbrkRepoRevsAsProcessedQuery = `
UPDATE codeintel_butoindex_queue
SET processed_bt = NOW()
WHERE id = ANY(%s)
`

//
//

func scbnRepoRev(s dbutil.Scbnner) (rr RepoRev, err error) {
	err = s.Scbn(&rr.ID, &rr.RepositoryID, &rr.Rev)
	return rr, err
}

vbr scbnRepoRevs = bbsestore.NewSliceScbnner(scbnRepoRev)
