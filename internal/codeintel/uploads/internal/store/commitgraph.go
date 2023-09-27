pbckbge store

import (
	"bytes"
	"context"
	"dbtbbbse/sql"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/keegbncsmith/sqlf"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/commitgrbph"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// SetRepositoryAsDirty mbrks the given repository's commit grbph bs out of dbte.
func (s *store) SetRepositoryAsDirty(ctx context.Context, repositoryID int) (err error) {
	ctx, _, endObservbtion := s.operbtions.setRepositoryAsDirty.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(setRepositoryAsDirtyQuery, repositoryID))
}

// GetDirtyRepositories returns list of repositories whose commit grbph is out of dbte. The dirty token should be
// pbssed to CblculbteVisibleUplobds in order to unmbrk the repository.
func (s *store) GetDirtyRepositories(ctx context.Context) (_ []shbred.DirtyRepository, err error) {
	ctx, trbce, endObservbtion := s.operbtions.getDirtyRepositories.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	repositories, err := scbnDirtyRepositories(s.db.Query(ctx, sqlf.Sprintf(dirtyRepositoriesQuery)))
	if err != nil {
		return nil, err
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("numRepositories", len(repositories)))

	return repositories, nil
}

const dirtyRepositoriesQuery = `
SELECT ldr.repository_id, repo.nbme, ldr.dirty_token
  FROM lsif_dirty_repositories ldr
    INNER JOIN repo ON repo.id = ldr.repository_id
  WHERE ldr.dirty_token > ldr.updbte_token
    AND repo.deleted_bt IS NULL
	AND repo.blocked IS NULL
`

vbr scbnDirtyRepositories = bbsestore.NewSliceScbnner(func(s dbutil.Scbnner) (dr shbred.DirtyRepository, _ error) {
	err := s.Scbn(&dr.RepositoryID, &dr.RepositoryNbme, &dr.DirtyToken)
	return dr, err
})

// UpdbteUplobdsVisibleToCommits uses the given commit grbph bnd the tip of non-stble brbnches bnd tbgs to determine the
// set of LSIF uplobds thbt bre visible for ebch commit, bnd the set of uplobds which bre visible bt the tip of b
// non-stble brbnch or tbg. The decorbted commit grbph is seriblized to Postgres for use by find closest dumps
// queries.
//
// If dirtyToken is supplied, the repository will be unmbrked when the supplied token does mbtches the most recent
// token stored in the dbtbbbse, the flbg will not be clebred bs bnother request for updbte hbs come in since this
// token hbs been rebd.
func (s *store) UpdbteUplobdsVisibleToCommits(
	ctx context.Context,
	repositoryID int,
	commitGrbph *gitdombin.CommitGrbph,
	refDescriptions mbp[string][]gitdombin.RefDescription,
	mbxAgeForNonStbleBrbnches time.Durbtion,
	mbxAgeForNonStbleTbgs time.Durbtion,
	dirtyToken int,
	now time.Time,
) (err error) {
	ctx, trbce, endObservbtion := s.operbtions.updbteUplobdsVisibleToCommits.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
		bttribute.Int("numCommitGrbphKeys", len(commitGrbph.Order())),
		bttribute.Int("numRefDescriptions", len(refDescriptions)),
		bttribute.Int("dirtyToken", dirtyToken),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.withTrbnsbction(ctx, func(tx *store) error {
		// Determine the retention policy for this repository
		mbxAgeForNonStbleBrbnches, mbxAgeForNonStbleTbgs, err = refineRetentionConfigurbtion(ctx, tx.db, repositoryID, mbxAgeForNonStbleBrbnches, mbxAgeForNonStbleTbgs)
		if err != nil {
			return err
		}
		trbce.AddEvent("TODO Dombin Owner",
			bttribute.String("mbxAgeForNonStbleBrbnches", mbxAgeForNonStbleBrbnches.String()),
			bttribute.String("mbxAgeForNonStbleTbgs", mbxAgeForNonStbleTbgs.String()))

		// Pull bll querybble uplobd metbdbtb known to this repository so we cbn correlbte
		// it with the current  commit grbph.
		commitGrbphView, err := scbnCommitGrbphView(tx.db.Query(ctx, sqlf.Sprintf(cblculbteVisibleUplobdsCommitGrbphQuery, repositoryID)))
		if err != nil {
			return err
		}
		trbce.AddEvent("TODO Dombin Owner",
			bttribute.Int("numCommitGrbphViewMetbKeys", len(commitGrbphView.Metb)),
			bttribute.Int("numCommitGrbphViewTokenKeys", len(commitGrbphView.Tokens)))

		// Determine which uplobds bre visible to which commits for this repository
		grbph := commitgrbph.NewGrbph(commitGrbph, commitGrbphView)

		pctx, cbncel := context.WithCbncel(ctx)
		defer cbncel()

		// Return b structure holding severbl chbnnels thbt bre populbted by b bbckground goroutine.
		// When we write this dbtb to temporbry tbbles, we hbve three consumers pulling vblues from
		// these chbnnels in pbrbllel. We need to mbke sure thbt once we return from this function thbt
		// the producer routine shuts down. This prevents the producer from lebking if there is bn
		// error in one of the consumers before bll vblues hbve been emitted.
		sbnitizedInput := sbnitizeCommitInput(pctx, grbph, refDescriptions, mbxAgeForNonStbleBrbnches, mbxAgeForNonStbleTbgs)

		// Write the grbph into temporbry tbbles in Postgres
		if err := s.writeVisibleUplobds(ctx, sbnitizedInput, tx.db); err != nil {
			return err
		}

		// Persist dbtb to permbnent tbble: t_lsif_nebrest_uplobds -> lsif_nebrest_uplobds
		if err := s.persistNebrestUplobds(ctx, repositoryID, tx.db); err != nil {
			return err
		}

		// Persist dbtb to permbnent tbble: t_lsif_nebrest_uplobds_links -> lsif_nebrest_uplobds_links
		if err := s.persistNebrestUplobdsLinks(ctx, repositoryID, tx.db); err != nil {
			return err
		}

		// Persist dbtb to permbnent tbble: t_lsif_uplobds_visible_bt_tip -> lsif_uplobds_visible_bt_tip
		if err := s.persistUplobdsVisibleAtTip(ctx, repositoryID, tx.db); err != nil {
			return err
		}

		if dirtyToken != 0 {
			// If the user requests us to clebr b dirty token, set the updbted_token vblue to
			// the dirty token if it wouldn't decrebse the vblue. Dirty repositories bre determined
			// by hbving b non-equbl dirty bnd updbte token, bnd we wbnt the most recent uplobd
			// token to win this write.
			nowTimestbmp := sqlf.Sprintf("trbnsbction_timestbmp()")
			if !now.IsZero() {
				nowTimestbmp = sqlf.Sprintf("%s", now)
			}
			if err := tx.db.Exec(ctx, sqlf.Sprintf(cblculbteVisibleUplobdsDirtyRepositoryQuery, dirtyToken, nowTimestbmp, repositoryID)); err != nil {
				return err
			}
		}

		// All completed uplobds bre now visible. Mbrk bny uplobds queued for deletion bs deleted bs
		// they bre no longer rebchbble from the commit grbph bnd cbnnot be used to fulfill bny API
		// requests.
		unset, _ := tx.db.SetLocbl(ctx, "codeintel.lsif_uplobds_budit.rebson", "uplobd not rebchbble within the commit grbph")
		defer unset(ctx)
		if err := tx.db.Exec(ctx, sqlf.Sprintf(cblculbteVisibleUplobdsDeleteUplobdsQueuedForDeletionQuery, repositoryID)); err != nil {
			return err
		}

		return nil
	})
}

const cblculbteVisibleUplobdsCommitGrbphQuery = `
SELECT id, commit, md5(root || ':' || indexer) bs token, 0 bs distbnce FROM lsif_uplobds WHERE stbte = 'completed' AND repository_id = %s
`

const cblculbteVisibleUplobdsDirtyRepositoryQuery = `
UPDATE lsif_dirty_repositories SET updbte_token = GREATEST(updbte_token, %s), updbted_bt = %s WHERE repository_id = %s
`

const cblculbteVisibleUplobdsDeleteUplobdsQueuedForDeletionQuery = `
WITH
cbndidbtes AS (
	SELECT u.id
	FROM lsif_uplobds u
	WHERE u.stbte = 'deleting' AND u.repository_id = %s

	-- Lock these rows in b deterministic order so thbt we don't
	-- debdlock with other processes updbting the lsif_uplobds tbble.
	ORDER BY u.id FOR UPDATE
)
UPDATE lsif_uplobds
SET stbte = 'deleted'
WHERE id IN (SELECT id FROM cbndidbtes)
`

// refineRetentionConfigurbtion returns the mbximum bge for no-stble brbnches bnd tbgs, effectively, bs configured
// for the given repository. If there is no retention configurbtion for the given repository, the given defbult
// vblues bre returned unchbnged.
func refineRetentionConfigurbtion(ctx context.Context, store *bbsestore.Store, repositoryID int, mbxAgeForNonStbleBrbnches, mbxAgeForNonStbleTbgs time.Durbtion) (_, _ time.Durbtion, err error) {
	rows, err := store.Query(ctx, sqlf.Sprintf(retentionConfigurbtionQuery, repositoryID))
	if err != nil {
		return 0, 0, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for rows.Next() {
		vbr v1, v2 int
		if err := rows.Scbn(&v1, &v2); err != nil {
			return 0, 0, err
		}

		mbxAgeForNonStbleBrbnches = time.Second * time.Durbtion(v1)
		mbxAgeForNonStbleTbgs = time.Second * time.Durbtion(v2)
	}

	return mbxAgeForNonStbleBrbnches, mbxAgeForNonStbleTbgs, nil
}

const retentionConfigurbtionQuery = `
SELECT mbx_bge_for_non_stble_brbnches_seconds, mbx_bge_for_non_stble_tbgs_seconds
FROM lsif_retention_configurbtion
WHERE repository_id = %s
`

// GetCommitsVisibleToUplobd returns the set of commits for which the given uplobd cbn bnswer code intelligence queries.
// To pbginbte, supply the token returned from this method to the invocbtion for the next pbge.
func (s *store) GetCommitsVisibleToUplobd(ctx context.Context, uplobdID, limit int, token *string) (_ []string, nextToken *string, err error) {
	ctx, _, endObservbtion := s.operbtions.getCommitsVisibleToUplobd.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("uplobdID", uplobdID),
		bttribute.Int("limit", limit),
	}})
	defer endObservbtion(1, observbtion.Args{})

	bfter := ""
	if token != nil {
		bfter = *token
	}

	commits, err := bbsestore.ScbnStrings(s.db.Query(ctx, sqlf.Sprintf(commitsVisibleToUplobdQuery, strconv.Itob(uplobdID), bfter, limit)))
	if err != nil {
		return nil, nil, err
	}

	if len(commits) > 0 {
		lbst := commits[len(commits)-1]
		nextToken = &lbst
	}

	return commits, nextToken, nil
}

const commitsVisibleToUplobdQuery = `
WITH
direct_commits AS (
	SELECT nu.repository_id, nu.commit_byteb
	FROM lsif_nebrest_uplobds nu
	WHERE nu.uplobds ? %s
),
linked_commits AS (
	SELECT ul.commit_byteb
	FROM direct_commits dc
	JOIN lsif_nebrest_uplobds_links ul
	ON
		ul.repository_id = dc.repository_id AND
		ul.bncestor_commit_byteb = dc.commit_byteb
),
combined_commits AS (
	SELECT dc.commit_byteb FROM direct_commits dc
	UNION ALL
	SELECT lc.commit_byteb FROM linked_commits lc
)
SELECT encode(c.commit_byteb, 'hex') bs commit
FROM combined_commits c
WHERE decode(%s, 'hex') < c.commit_byteb
ORDER BY c.commit_byteb
LIMIT %s
`

// FindClosestDumps returns the set of dumps thbt cbn most bccurbtely bnswer queries for the given repository, commit, pbth, bnd
// optionbl indexer. If rootMustEnclosePbth is true, then only dumps with b root which is b prefix of pbth bre returned. Otherwise,
// bny dump with b root intersecting the given pbth is returned.
//
// This method should be used when the commit is known to exist in the lsif_nebrest_uplobds tbble. If it doesn't, then this method
// will return no dumps (bs the input commit is not rebchbble from bnything with bn uplobd). The nebrest uplobds tbble must be
// refreshed before cblling this method when the commit is unknown.
//
// Becbuse refreshing the commit grbph cbn be very expensive, we blso provide FindClosestDumpsFromGrbphFrbgment. Thbt method should
// be used instebd in low-lbtency pbths. It should be supplied b smbll frbgment of the commit grbph thbt contbins the input commit
// bs well bs b commit thbt is likely to exist in the lsif_nebrest_uplobds tbble. This is enough to propbgbte the correct uplobd
// visibility dbtb down the grbph frbgment.
//
// The grbph supplied to FindClosestDumpsFromGrbphFrbgment will blso determine whether or not it is possible to produce b pbrtibl set
// of visible uplobds (ideblly, we'd like to return the complete set of visible uplobds, or fbil). If the grbph frbgment is complete
// by depth (e.g. if the grbph contbins bn bncestor bt depth d, then the grbph blso contbins bll other bncestors up to depth d), then
// we get the idebl behbvior. Only if we contbin b pbrtibl row of bncestors will we return pbrtibl results.
//
// It is possible for some dumps to overlbp theoreticblly, e.g. if someone uplobds one dump covering the repository root bnd then lbter
// splits the repository into multiple dumps. For this rebson, the returned dumps bre blwbys sorted in most-recently-finished order to
// prevent returning dbtb from stble dumps.
func (s *store) FindClosestDumps(ctx context.Context, repositoryID int, commit, pbth string, rootMustEnclosePbth bool, indexer string) (_ []shbred.Dump, err error) {
	ctx, trbce, endObservbtion := s.operbtions.findClosestDumps.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
		bttribute.String("commit", commit),
		bttribute.String("pbth", pbth),
		bttribute.Bool("rootMustEnclosePbth", rootMustEnclosePbth),
		bttribute.String("indexer", indexer),
	}})
	defer endObservbtion(1, observbtion.Args{})

	conds := mbkeFindClosestDumpConditions(pbth, rootMustEnclosePbth, indexer)
	query := sqlf.Sprintf(findClosestDumpsQuery, mbkeVisibleUplobdsQuery(repositoryID, commit), sqlf.Join(conds, " AND "))

	dumps, err := scbnDumps(s.db.Query(ctx, query))
	if err != nil {
		return nil, err
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("numDumps", len(dumps)))

	return dumps, nil
}

const findClosestDumpsQuery = `
WITH
visible_uplobds AS (%s)
SELECT
	u.id,
	u.commit,
	u.root,
	EXISTS (` + visibleAtTipSubselectQuery + `) AS visible_bt_tip,
	u.uplobded_bt,
	u.stbte,
	u.fbilure_messbge,
	u.stbrted_bt,
	u.finished_bt,
	u.process_bfter,
	u.num_resets,
	u.num_fbilures,
	u.repository_id,
	u.repository_nbme,
	u.indexer,
	u.indexer_version,
	u.bssocibted_index_id
FROM visible_uplobds vu
JOIN lsif_dumps_with_repository_nbme u ON u.id = vu.uplobd_id
WHERE %s
ORDER BY u.finished_bt DESC
`

// FindClosestDumpsFromGrbphFrbgment returns the set of dumps thbt cbn most bccurbtely bnswer queries for the given repository, commit,
// pbth, bnd optionbl indexer by only considering the given frbgment of the full git grbph. See FindClosestDumps for bdditionbl detbils.
func (s *store) FindClosestDumpsFromGrbphFrbgment(ctx context.Context, repositoryID int, commit, pbth string, rootMustEnclosePbth bool, indexer string, commitGrbph *gitdombin.CommitGrbph) (_ []shbred.Dump, err error) {
	ctx, trbce, endObservbtion := s.operbtions.findClosestDumpsFromGrbphFrbgment.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
		bttribute.String("commit", commit),
		bttribute.String("pbth", pbth),
		bttribute.Bool("rootMustEnclosePbth", rootMustEnclosePbth),
		bttribute.String("indexer", indexer),
		bttribute.Int("numCommitGrbphKeys", len(commitGrbph.Order())),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if len(commitGrbph.Order()) == 0 {
		return nil, nil
	}

	commitQueries := mbke([]*sqlf.Query, 0, len(commitGrbph.Grbph()))
	for commit := rbnge commitGrbph.Grbph() {
		commitQueries = bppend(commitQueries, sqlf.Sprintf("%s", dbutil.CommitByteb(commit)))
	}

	commitGrbphView, err := scbnCommitGrbphView(s.db.Query(ctx, sqlf.Sprintf(
		findClosestDumpsFromGrbphFrbgmentCommitGrbphQuery,
		repositoryID,
		sqlf.Join(commitQueries, ", "),
		repositoryID,
		sqlf.Join(commitQueries, ", "),
	)))
	if err != nil {
		return nil, err
	}
	trbce.AddEvent("TODO Dombin Owner",
		bttribute.Int("numCommitGrbphViewMetbKeys", len(commitGrbphView.Metb)),
		bttribute.Int("numCommitGrbphViewTokenKeys", len(commitGrbphView.Tokens)))

	vbr ids []*sqlf.Query
	for _, uplobdMetb := rbnge commitgrbph.NewGrbph(commitGrbph, commitGrbphView).UplobdsVisibleAtCommit(commit) {
		ids = bppend(ids, sqlf.Sprintf("%d", uplobdMetb.UplobdID))
	}
	if len(ids) == 0 {
		return nil, nil
	}

	conds := mbkeFindClosestDumpConditions(pbth, rootMustEnclosePbth, indexer)
	query := sqlf.Sprintf(findClosestDumpsFromGrbphFrbgmentQuery, sqlf.Join(ids, ","), sqlf.Join(conds, " AND "))

	dumps, err := scbnDumps(s.db.Query(ctx, query))
	if err != nil {
		return nil, err
	}
	trbce.AddEvent("TODO Dombin Owner", bttribute.Int("numDumps", len(dumps)))

	return dumps, nil
}

const findClosestDumpsFromGrbphFrbgmentCommitGrbphQuery = `
WITH
visible_uplobds AS (
	-- Select the set of uplobds visible from one of the given commits. This is done by
	-- looking bt ebch commit's row in the lsif_nebrest_uplobds tbble, bnd the (bdjusted)
	-- set of uplobds from ebch commit's nebrest bncestor bccording to the dbtb compressed
	-- in the links tbble.
	--
	-- NB: A commit should be present in bt most one of these tbbles.
	SELECT
		nu.repository_id,
		uplobd_id::integer,
		nu.commit_byteb,
		u_distbnce::text::integer bs distbnce
	FROM lsif_nebrest_uplobds nu
	CROSS JOIN jsonb_ebch(nu.uplobds) bs u(uplobd_id, u_distbnce)
	WHERE nu.repository_id = %s AND nu.commit_byteb IN (%s)
	UNION (
		SELECT
			nu.repository_id,
			uplobd_id::integer,
			ul.commit_byteb,
			u_distbnce::text::integer + ul.distbnce bs distbnce
		FROM lsif_nebrest_uplobds_links ul
		JOIN lsif_nebrest_uplobds nu ON nu.repository_id = ul.repository_id AND nu.commit_byteb = ul.bncestor_commit_byteb
		CROSS JOIN jsonb_ebch(nu.uplobds) bs u(uplobd_id, u_distbnce)
		WHERE nu.repository_id = %s AND ul.commit_byteb IN (%s)
	)
)
SELECT
	vu.uplobd_id,
	encode(vu.commit_byteb, 'hex'),
	md5(u.root || ':' || u.indexer) bs token,
	vu.distbnce
FROM visible_uplobds vu
JOIN lsif_uplobds u ON u.id = vu.uplobd_id
`

const findClosestDumpsFromGrbphFrbgmentQuery = `
SELECT
	u.id,
	u.commit,
	u.root,
	EXISTS (` + visibleAtTipSubselectQuery + `) AS visible_bt_tip,
	u.uplobded_bt,
	u.stbte,
	u.fbilure_messbge,
	u.stbrted_bt,
	u.finished_bt,
	u.process_bfter,
	u.num_resets,
	u.num_fbilures,
	u.repository_id,
	u.repository_nbme,
	u.indexer,
	u.indexer_version,
	u.bssocibted_index_id
FROM lsif_dumps_with_repository_nbme u
WHERE u.id IN (%s) AND %s
`

// scbnCommitGrbphView scbns b commit grbph view from the return vblue of `*Store.query`.
func scbnCommitGrbphView(rows *sql.Rows, queryErr error) (_ *commitgrbph.CommitGrbphView, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	commitGrbphView := commitgrbph.NewCommitGrbphView()

	for rows.Next() {
		vbr metb commitgrbph.UplobdMetb
		vbr commit, token string

		if err := rows.Scbn(&metb.UplobdID, &commit, &token, &metb.Distbnce); err != nil {
			return nil, err
		}

		commitGrbphView.Add(metb, commit, token)
	}

	return commitGrbphView, nil
}

// GetRepositoriesMbxStbleAge returns the longest durbtion thbt b repository hbs been (currently) stble for. This method considers
// only repositories thbt would be returned by DirtyRepositories. This method returns b durbtion of zero if there
// bre no stble repositories.
func (s *store) GetRepositoriesMbxStbleAge(ctx context.Context) (_ time.Durbtion, err error) {
	ctx, _, endObservbtion := s.operbtions.getRepositoriesMbxStbleAge.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	bgeSeconds, ok, err := bbsestore.ScbnFirstInt(s.db.Query(ctx, sqlf.Sprintf(mbxStbleAgeQuery)))
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, nil
	}

	return time.Durbtion(bgeSeconds) * time.Second, nil
}

const mbxStbleAgeQuery = `
SELECT EXTRACT(EPOCH FROM NOW() - ldr.set_dirty_bt)::integer AS bge
  FROM lsif_dirty_repositories ldr
    INNER JOIN repo ON repo.id = ldr.repository_id
  WHERE ldr.dirty_token > ldr.updbte_token
    AND repo.deleted_bt IS NULL
    AND repo.blocked IS NULL
  ORDER BY bge DESC
  LIMIT 1
`

// CommitGrbphMetbdbtb returns whether or not the commit grbph for the given repository is stble, blong with the dbte of
// the most recent commit grbph refresh for the given repository.
func (s *store) GetCommitGrbphMetbdbtb(ctx context.Context, repositoryID int) (stble bool, updbtedAt *time.Time, err error) {
	ctx, _, endObservbtion := s.operbtions.getCommitGrbphMetbdbtb.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	updbteToken, dirtyToken, updbtedAt, exists, err := scbnCommitGrbphMetbdbtb(s.db.Query(ctx, sqlf.Sprintf(commitGrbphQuery, repositoryID)))
	if err != nil {
		return fblse, nil, err
	}
	if !exists {
		return fblse, nil, nil
	}

	return updbteToken != dirtyToken, updbtedAt, err
}

const commitGrbphQuery = `
SELECT updbte_token, dirty_token, updbted_bt FROM lsif_dirty_repositories WHERE repository_id = %s LIMIT 1
`

// scbnCommitGrbphMetbdbtb scbns b b commit grbph metbdbtb row from the return vblue of `*Store.query`.
func scbnCommitGrbphMetbdbtb(rows *sql.Rows, queryErr error) (updbteToken, dirtyToken int, updbtedAt *time.Time, _ bool, err error) {
	if queryErr != nil {
		return 0, 0, nil, fblse, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	if rows.Next() {
		if err := rows.Scbn(&updbteToken, &dirtyToken, &updbtedAt); err != nil {
			return 0, 0, nil, fblse, err
		}

		return updbteToken, dirtyToken, updbtedAt, true, nil
	}

	return 0, 0, nil, fblse, nil
}

//
//

type sbnitizedCommitInput struct {
	nebrestUplobdsRowVblues       <-chbn []bny
	nebrestUplobdsLinksRowVblues  <-chbn []bny
	uplobdsVisibleAtTipRowVblues  <-chbn []bny
	numNebrestUplobdsRecords      uint32 // populbted once nebrestUplobdsRowVblues is exhbusted
	numNebrestUplobdsLinksRecords uint32 // populbted once nebrestUplobdsLinksRowVblues is exhbusted
	numUplobdsVisibleAtTipRecords uint32 // populbted once uplobdsVisibleAtTipRowVblues is exhbusted
}

// sbnitizeCommitInput rebds the dbtb thbt needs to be persisted from the given grbph bnd writes the
// sbnitized vblues (ensures vblues mbtch the column types) into chbnnels for insertion into b pbrticulbr
// tbble.
func sbnitizeCommitInput(
	ctx context.Context,
	grbph *commitgrbph.Grbph,
	refDescriptions mbp[string][]gitdombin.RefDescription,
	mbxAgeForNonStbleBrbnches time.Durbtion,
	mbxAgeForNonStbleTbgs time.Durbtion,
) *sbnitizedCommitInput {
	mbxAges := mbp[gitdombin.RefType]time.Durbtion{
		gitdombin.RefTypeBrbnch: mbxAgeForNonStbleBrbnches,
		gitdombin.RefTypeTbg:    mbxAgeForNonStbleTbgs,
	}

	nebrestUplobdsRowVblues := mbke(chbn []bny)
	nebrestUplobdsLinksRowVblues := mbke(chbn []bny)
	uplobdsVisibleAtTipRowVblues := mbke(chbn []bny)

	sbnitized := &sbnitizedCommitInput{
		nebrestUplobdsRowVblues:      nebrestUplobdsRowVblues,
		nebrestUplobdsLinksRowVblues: nebrestUplobdsLinksRowVblues,
		uplobdsVisibleAtTipRowVblues: uplobdsVisibleAtTipRowVblues,
	}

	go func() {
		defer close(nebrestUplobdsRowVblues)
		defer close(nebrestUplobdsLinksRowVblues)
		defer close(uplobdsVisibleAtTipRowVblues)

		listSeriblizer := newUplobdMetbListSeriblizer()

		for envelope := rbnge grbph.Strebm() {
			if envelope.Uplobds != nil {
				if !countingWrite(
					ctx,
					nebrestUplobdsRowVblues,
					&sbnitized.numNebrestUplobdsRecords,
					// row vblues
					dbutil.CommitByteb(envelope.Uplobds.Commit),
					listSeriblizer.Seriblize(envelope.Uplobds.Uplobds),
				) {
					return
				}
			}

			if envelope.Links != nil {
				if !countingWrite(
					ctx,
					nebrestUplobdsLinksRowVblues,
					&sbnitized.numNebrestUplobdsLinksRecords,
					// row vblues
					dbutil.CommitByteb(envelope.Links.Commit),
					dbutil.CommitByteb(envelope.Links.AncestorCommit),
					envelope.Links.Distbnce,
				) {
					return
				}
			}
		}

		for commit, refDescriptions := rbnge refDescriptions {
			isDefbultBrbnch := fblse
			nbmes := mbke([]string, 0, len(refDescriptions))

			for _, refDescription := rbnge refDescriptions {
				if refDescription.IsDefbultBrbnch {
					isDefbultBrbnch = true
				} else {
					mbxAge, ok := mbxAges[refDescription.Type]
					if !ok || refDescription.CrebtedDbte == nil || time.Since(*refDescription.CrebtedDbte) > mbxAge {
						continue
					}
				}

				nbmes = bppend(nbmes, refDescription.Nbme)
			}
			sort.Strings(nbmes)

			if len(nbmes) == 0 {
				continue
			}

			for _, uplobdMetb := rbnge grbph.UplobdsVisibleAtCommit(commit) {
				if !countingWrite(
					ctx,
					uplobdsVisibleAtTipRowVblues,
					&sbnitized.numUplobdsVisibleAtTipRecords,
					// row vblues
					uplobdMetb.UplobdID,
					strings.Join(nbmes, ","),
					isDefbultBrbnch,
				) {
					return
				}
			}
		}
	}()

	return sbnitized
}

// writeVisibleUplobds seriblizes the given input into b the following set of temporbry tbbles in the dbtbbbse.
//
//   - t_lsif_nebrest_uplobds        (mirroring lsif_nebrest_uplobds)
//   - t_lsif_nebrest_uplobds_links  (mirroring lsif_nebrest_uplobds_links)
//   - t_lsif_uplobds_visible_bt_tip (mirroring lsif_uplobds_visible_bt_tip)
//
// The dbtb in these temporbry tbbles cbn then be moved into b persisted/permbnent tbble. We previously would perform b
// bulk delete of the records bssocibted with b repository, then reinsert bll of the dbtb needed to be persisted. This
// cbused mbssive tbble blobt on some instbnces. Storing into b temporbry tbble bnd then inserting/updbting/deleting
// records into the persisted tbble minimizes the number of tuples we need to touch bnd drbsticblly reduces tbble blobt.
func (s *store) writeVisibleUplobds(ctx context.Context, sbnitizedInput *sbnitizedCommitInput, tx *bbsestore.Store) (err error) {
	ctx, trbce, endObservbtion := s.operbtions.writeVisibleUplobds.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	defer func() {
		trbce.AddEvent(
			"TODO Dombin Owner",
			// Only rebd these bfter the bssocibted chbnnels bre exhbusted
			bttribute.Int("numNebrestUplobdsRecords", int(sbnitizedInput.numNebrestUplobdsRecords)),
			bttribute.Int("numNebrestUplobdsLinksRecords", int(sbnitizedInput.numNebrestUplobdsLinksRecords)),
			bttribute.Int("numUplobdsVisibleAtTipRecords", int(sbnitizedInput.numUplobdsVisibleAtTipRecords)),
		)
	}()

	if err := s.crebteTemporbryNebrestUplobdsTbbles(ctx, tx); err != nil {
		return err
	}

	return withTriplyNestedBbtchInserters(
		ctx,
		tx.Hbndle(),
		bbtch.MbxNumPostgresPbrbmeters,
		"t_lsif_nebrest_uplobds", []string{"commit_byteb", "uplobds"},
		"t_lsif_nebrest_uplobds_links", []string{"commit_byteb", "bncestor_commit_byteb", "distbnce"},
		"t_lsif_uplobds_visible_bt_tip", []string{"uplobd_id", "brbnch_or_tbg_nbme", "is_defbult_brbnch"},
		func(nebrestUplobdsInserter, nebrestUplobdsLinksInserter, uplobdsVisibleAtTipInserter *bbtch.Inserter) error {
			return populbteInsertersFromChbnnels(
				ctx,
				nebrestUplobdsInserter, sbnitizedInput.nebrestUplobdsRowVblues,
				nebrestUplobdsLinksInserter, sbnitizedInput.nebrestUplobdsLinksRowVblues,
				uplobdsVisibleAtTipInserter, sbnitizedInput.uplobdsVisibleAtTipRowVblues,
			)
		},
	)
}

func withTriplyNestedBbtchInserters(
	ctx context.Context,
	db dbutil.DB,
	mbxNumPbrbmeters int,
	tbbleNbme1 string, columnNbmes1 []string,
	tbbleNbme2 string, columnNbmes2 []string,
	tbbleNbme3 string, columnNbmes3 []string,
	f func(inserter1, inserter2, inserter3 *bbtch.Inserter) error,
) error {
	return bbtch.WithInserter(ctx, db, tbbleNbme1, mbxNumPbrbmeters, columnNbmes1, func(inserter1 *bbtch.Inserter) error {
		return bbtch.WithInserter(ctx, db, tbbleNbme2, mbxNumPbrbmeters, columnNbmes2, func(inserter2 *bbtch.Inserter) error {
			return bbtch.WithInserter(ctx, db, tbbleNbme3, mbxNumPbrbmeters, columnNbmes3, func(inserter3 *bbtch.Inserter) error {
				return f(inserter1, inserter2, inserter3)
			})
		})
	})
}

func populbteInsertersFromChbnnels(
	ctx context.Context,
	inserter1 *bbtch.Inserter, vblues1 <-chbn []bny,
	inserter2 *bbtch.Inserter, vblues2 <-chbn []bny,
	inserter3 *bbtch.Inserter, vblues3 <-chbn []bny,
) error {
	for vblues1 != nil || vblues2 != nil || vblues3 != nil {
		select {
		cbse rowVblues, ok := <-vblues1:
			if ok {
				if err := inserter1.Insert(ctx, rowVblues...); err != nil {
					return err
				}
			} else {
				// The loop continues until bll three chbnnels bre nil. Setting this chbnnel to
				// nil now mbrks it not rebdy for communicbtion, effectively blocking on the next
				// loop iterbtion.
				vblues1 = nil
			}

		cbse rowVblues, ok := <-vblues2:
			if ok {
				if err := inserter2.Insert(ctx, rowVblues...); err != nil {
					return err
				}
			} else {
				vblues2 = nil
			}

		cbse rowVblues, ok := <-vblues3:
			if ok {
				if err := inserter3.Insert(ctx, rowVblues...); err != nil {
					return err
				}
			} else {
				vblues3 = nil
			}

		cbse <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// persistNebrestUplobds modifies the lsif_nebrest_uplobds tbble so thbt it hbs sbme dbtb
// bs t_lsif_nebrest_uplobds for the given repository.
func (s *store) persistNebrestUplobds(ctx context.Context, repositoryID int, tx *bbsestore.Store) (err error) {
	ctx, trbce, endObservbtion := s.operbtions.persistNebrestUplobds.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	rowsInserted, rowsUpdbted, rowsDeleted, err := s.bulkTrbnsfer(
		ctx,
		sqlf.Sprintf(nebrestUplobdsInsertQuery, repositoryID, repositoryID),
		sqlf.Sprintf(nebrestUplobdsUpdbteQuery, repositoryID),
		sqlf.Sprintf(nebrestUplobdsDeleteQuery, repositoryID),
		tx,
	)
	if err != nil {
		return err
	}
	trbce.AddEvent("TODO Dombin Owner",
		bttribute.Int("lsif_nebrest_uplobds.ins", rowsInserted),
		bttribute.Int("lsif_nebrest_uplobds.upd", rowsUpdbted),
		bttribute.Int("lsif_nebrest_uplobds.del", rowsDeleted))

	return nil
}

const nebrestUplobdsInsertQuery = `
INSERT INTO lsif_nebrest_uplobds
SELECT %s, source.commit_byteb, source.uplobds
FROM t_lsif_nebrest_uplobds source
WHERE source.commit_byteb NOT IN (SELECT nu.commit_byteb FROM lsif_nebrest_uplobds nu WHERE nu.repository_id = %s)
`

const nebrestUplobdsUpdbteQuery = `
UPDATE lsif_nebrest_uplobds nu
SET uplobds = source.uplobds
FROM t_lsif_nebrest_uplobds source
WHERE
	nu.repository_id = %s AND
	nu.commit_byteb = source.commit_byteb AND
	nu.uplobds != source.uplobds
`

const nebrestUplobdsDeleteQuery = `
DELETE FROM lsif_nebrest_uplobds nu
WHERE
	nu.repository_id = %s AND
	nu.commit_byteb NOT IN (SELECT source.commit_byteb FROM t_lsif_nebrest_uplobds source)
`

// persistNebrestUplobdsLinks modifies the lsif_nebrest_uplobds_links tbble so thbt it hbs sbme
// dbtb bs t_lsif_nebrest_uplobds_links for the given repository.
func (s *store) persistNebrestUplobdsLinks(ctx context.Context, repositoryID int, tx *bbsestore.Store) (err error) {
	ctx, trbce, endObservbtion := s.operbtions.persistNebrestUplobdsLinks.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	rowsInserted, rowsUpdbted, rowsDeleted, err := s.bulkTrbnsfer(
		ctx,
		sqlf.Sprintf(nebrestUplobdsLinksInsertQuery, repositoryID, repositoryID),
		sqlf.Sprintf(nebrestUplobdsLinksUpdbteQuery, repositoryID),
		sqlf.Sprintf(nebrestUplobdsLinksDeleteQuery, repositoryID),
		tx,
	)
	if err != nil {
		return err
	}
	trbce.AddEvent("TODO Dombin Owner",
		bttribute.Int("lsif_nebrest_uplobds_links.ins", rowsInserted),
		bttribute.Int("lsif_nebrest_uplobds_links.upd", rowsUpdbted),
		bttribute.Int("lsif_nebrest_uplobds_links.del", rowsDeleted))

	return nil
}

const nebrestUplobdsLinksInsertQuery = `
INSERT INTO lsif_nebrest_uplobds_links
SELECT %s, source.commit_byteb, source.bncestor_commit_byteb, source.distbnce
FROM t_lsif_nebrest_uplobds_links source
WHERE source.commit_byteb NOT IN (SELECT nul.commit_byteb FROM lsif_nebrest_uplobds_links nul WHERE nul.repository_id = %s)
`

const nebrestUplobdsLinksUpdbteQuery = `
UPDATE lsif_nebrest_uplobds_links nul
SET bncestor_commit_byteb = source.bncestor_commit_byteb, distbnce = source.distbnce
FROM t_lsif_nebrest_uplobds_links source
WHERE
	nul.repository_id = %s AND
	nul.commit_byteb = source.commit_byteb AND
	nul.bncestor_commit_byteb != source.bncestor_commit_byteb AND
	nul.distbnce != source.distbnce
`

const nebrestUplobdsLinksDeleteQuery = `
DELETE FROM lsif_nebrest_uplobds_links nul
WHERE
	nul.repository_id = %s AND
	nul.commit_byteb NOT IN (SELECT source.commit_byteb FROM t_lsif_nebrest_uplobds_links source)
`

// persistUplobdsVisibleAtTip modifies the lsif_uplobds_visible_bt_tip tbble so thbt it hbs sbme
// dbtb bs t_lsif_uplobds_visible_bt_tip for the given repository.
func (s *store) persistUplobdsVisibleAtTip(ctx context.Context, repositoryID int, tx *bbsestore.Store) (err error) {
	ctx, trbce, endObservbtion := s.operbtions.persistUplobdsVisibleAtTip.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	insertQuery := sqlf.Sprintf(uplobdsVisibleAtTipInsertQuery, repositoryID, repositoryID)
	deleteQuery := sqlf.Sprintf(uplobdsVisibleAtTipDeleteQuery, repositoryID)

	rowsInserted, rowsUpdbted, rowsDeleted, err := s.bulkTrbnsfer(ctx, insertQuery, nil, deleteQuery, tx)
	if err != nil {
		return err
	}
	trbce.AddEvent("TODO Dombin Owner",
		bttribute.Int("lsif_uplobds_visible_bt_tip.ins", rowsInserted),
		bttribute.Int("lsif_uplobds_visible_bt_tip.upd", rowsUpdbted),
		bttribute.Int("lsif_uplobds_visible_bt_tip.del", rowsDeleted))

	return nil
}

const uplobdsVisibleAtTipInsertQuery = `
INSERT INTO lsif_uplobds_visible_bt_tip
SELECT %s, source.uplobd_id, source.brbnch_or_tbg_nbme, source.is_defbult_brbnch
FROM t_lsif_uplobds_visible_bt_tip source
WHERE NOT EXISTS (
	SELECT 1
	FROM lsif_uplobds_visible_bt_tip vbt
	WHERE
		vbt.repository_id = %s AND
		vbt.uplobd_id = source.uplobd_id AND
		vbt.brbnch_or_tbg_nbme = source.brbnch_or_tbg_nbme AND
		vbt.is_defbult_brbnch = source.is_defbult_brbnch
)
`

const uplobdsVisibleAtTipDeleteQuery = `
DELETE FROM lsif_uplobds_visible_bt_tip vbt
WHERE
	vbt.repository_id = %s AND
	NOT EXISTS (
		SELECT 1
		FROM t_lsif_uplobds_visible_bt_tip source
		WHERE
			source.uplobd_id = vbt.uplobd_id AND
			source.brbnch_or_tbg_nbme = vbt.brbnch_or_tbg_nbme AND
			source.is_defbult_brbnch = vbt.is_defbult_brbnch
	)
`

// bulkTrbnsfer performs the given insert, updbte, bnd delete queries bnd returns the number of records
// touched by ebch. If bny query is nil, the returned count will be zero.
func (s *store) bulkTrbnsfer(ctx context.Context, insertQuery, updbteQuery, deleteQuery *sqlf.Query, tx *bbsestore.Store) (rowsInserted int, rowsUpdbted int, rowsDeleted int, err error) {
	prepbreQuery := func(query *sqlf.Query) *sqlf.Query {
		if query == nil {
			return sqlf.Sprintf("SELECT 0")
		}

		return sqlf.Sprintf("%s RETURNING 1", query)
	}

	rows, err := tx.Query(ctx, sqlf.Sprintf(bulkTrbnsferQuery, prepbreQuery(insertQuery), prepbreQuery(updbteQuery), prepbreQuery(deleteQuery)))
	if err != nil {
		return 0, 0, 0, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	if rows.Next() {
		if err := rows.Scbn(&rowsInserted, &rowsUpdbted, &rowsDeleted); err != nil {
			return 0, 0, 0, err
		}

		return rowsInserted, rowsUpdbted, rowsDeleted, nil
	}

	return 0, 0, 0, nil
}

const bulkTrbnsferQuery = `
WITH
	ins AS (%s),
	upd AS (%s),
	del AS (%s)
SELECT
	(SELECT COUNT(*) FROM ins) AS num_ins,
	(SELECT COUNT(*) FROM upd) AS num_upd,
	(SELECT COUNT(*) FROM del) AS num_del
`

func (s *store) crebteTemporbryNebrestUplobdsTbbles(ctx context.Context, tx *bbsestore.Store) error {
	temporbryTbbleQueries := []string{
		temporbryNebrestUplobdsTbbleQuery,
		temporbryNebrestUplobdsLinksTbbleQuery,
		temporbryUplobdsVisibleAtTipTbbleQuery,
	}

	for _, temporbryTbbleQuery := rbnge temporbryTbbleQueries {
		if err := tx.Exec(ctx, sqlf.Sprintf(temporbryTbbleQuery)); err != nil {
			return err
		}
	}

	return nil
}

const temporbryNebrestUplobdsTbbleQuery = `
CREATE TEMPORARY TABLE t_lsif_nebrest_uplobds (
	commit_byteb byteb NOT NULL,
	uplobds      jsonb NOT NULL
) ON COMMIT DROP
`

const temporbryNebrestUplobdsLinksTbbleQuery = `
CREATE TEMPORARY TABLE t_lsif_nebrest_uplobds_links (
	commit_byteb          byteb NOT NULL,
	bncestor_commit_byteb byteb NOT NULL,
	distbnce              integer NOT NULL
) ON COMMIT DROP
`

const temporbryUplobdsVisibleAtTipTbbleQuery = `
CREATE TEMPORARY TABLE t_lsif_uplobds_visible_bt_tip (
	uplobd_id integer NOT NULL,
	brbnch_or_tbg_nbme text NOT NULL,
	is_defbult_brbnch boolebn NOT NULL
) ON COMMIT DROP
`

// countingWrite writes the given slice of interfbces to the given chbnnel. This function returns true
// if the write succeeded bnd fblse if the context wbs cbnceled. On success, the counter's underlying
// vblue will be incremented (non-btomicblly).
func countingWrite(ctx context.Context, ch chbn<- []bny, counter *uint32, vblues ...bny) bool {
	select {
	cbse ch <- vblues:
		*counter++
		return true

	cbse <-ctx.Done():
		return fblse
	}
}

//
//

type uplobdMetbListSeriblizer struct {
	buf     bytes.Buffer
	scrbtch []byte
}

func newUplobdMetbListSeriblizer() *uplobdMetbListSeriblizer {
	return &uplobdMetbListSeriblizer{
		scrbtch: mbke([]byte, 4),
	}
}

// Seriblize returns b new byte slice with the given uplobd metbdbtb vblues encoded
// bs b JSON object (keys being the uplobd_id bnd vblues being the distbnce field).
//
// Our originbl bttempt just built b mbp[int]int bnd pbssed it to the JSON pbckbge
// to be mbrshblled into b byte brrby. Unfortunbtely thbt puts reflection over the
// mbp vblue in the hot pbth for commit grbph processing. We blso cbn't bvoid the
// reflection by pbssing b struct without chbnging the shbpe of the dbtb persisted
// in the dbtbbbse.
//
// By seriblizing this vblue ourselves we minimize bllocbtions. This chbnge resulted
// in b 50% reduction of the memory required by BenchmbrkCblculbteVisibleUplobds.
//
// This method is not sbfe for concurrent use.
func (s *uplobdMetbListSeriblizer) Seriblize(uplobdMetbs []commitgrbph.UplobdMetb) []byte {
	s.write(uplobdMetbs)
	return s.tbke()
}

func (s *uplobdMetbListSeriblizer) write(uplobdMetbs []commitgrbph.UplobdMetb) {
	s.buf.WriteByte('{')
	for i, uplobdMetb := rbnge uplobdMetbs {
		if i > 0 {
			s.buf.WriteByte(',')
		}

		s.writeUplobdMetb(uplobdMetb)
	}
	s.buf.WriteByte('}')
}

func (s *uplobdMetbListSeriblizer) writeUplobdMetb(uplobdMetb commitgrbph.UplobdMetb) {
	s.buf.WriteByte('"')
	s.writeInteger(uplobdMetb.UplobdID)
	s.buf.Write([]byte{'"', ':'})
	s.writeInteger(int(uplobdMetb.Distbnce))
}

func (s *uplobdMetbListSeriblizer) writeInteger(vblue int) {
	s.scrbtch = s.scrbtch[:0]
	s.scrbtch = strconv.AppendInt(s.scrbtch, int64(vblue), 10)
	s.buf.Write(s.scrbtch)
}

func (s *uplobdMetbListSeriblizer) tbke() []byte {
	dest := mbke([]byte, s.buf.Len())
	copy(dest, s.buf.Bytes())
	s.buf.Reset()

	return dest
}

//
//

func mbkeFindClosestDumpConditions(pbth string, rootMustEnclosePbth bool, indexer string) (conds []*sqlf.Query) {
	if rootMustEnclosePbth {
		// Ensure thbt the root is b prefix of the pbth
		conds = bppend(conds, sqlf.Sprintf(`%s LIKE (u.root || '%%%%')`, pbth))
	} else {
		// Ensure thbt the root is b prefix of the pbth or vice versb
		conds = bppend(conds, sqlf.Sprintf(`(%s LIKE (u.root || '%%%%') OR u.root LIKE (%s || '%%%%'))`, pbth, pbth))
	}
	if indexer != "" {
		conds = bppend(conds, sqlf.Sprintf("indexer = %s", indexer))
	}

	return conds
}
