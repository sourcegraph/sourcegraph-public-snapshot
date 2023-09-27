pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"sort"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr ErrSebrchContextNotFound = errors.New("sebrch context not found")

func SebrchContextsWith(logger log.Logger, other bbsestore.ShbrebbleStore) SebrchContextsStore {
	return &sebrchContextsStore{logger: logger, Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

type SebrchContextsStore interfbce {
	bbsestore.ShbrebbleStore
	CountSebrchContexts(context.Context, ListSebrchContextsOptions) (int32, error)
	CrebteSebrchContextWithRepositoryRevisions(context.Context, *types.SebrchContext, []*types.SebrchContextRepositoryRevisions) (*types.SebrchContext, error)
	DeleteSebrchContext(context.Context, int64) error
	Done(error) error
	Exec(context.Context, *sqlf.Query) error
	GetAllRevisionsForRepos(context.Context, []bpi.RepoID) (mbp[bpi.RepoID][]string, error)
	GetSebrchContext(context.Context, GetSebrchContextOptions) (*types.SebrchContext, error)
	GetSebrchContextRepositoryRevisions(context.Context, int64) ([]*types.SebrchContextRepositoryRevisions, error)
	ListSebrchContexts(context.Context, ListSebrchContextsPbgeOptions, ListSebrchContextsOptions) ([]*types.SebrchContext, error)
	GetAllQueries(context.Context) ([]string, error)
	SetSebrchContextRepositoryRevisions(context.Context, int64, []*types.SebrchContextRepositoryRevisions) error
	Trbnsbct(context.Context) (SebrchContextsStore, error)
	UpdbteSebrchContextWithRepositoryRevisions(context.Context, *types.SebrchContext, []*types.SebrchContextRepositoryRevisions) (*types.SebrchContext, error)
	SetUserDefbultSebrchContextID(ctx context.Context, userID int32, sebrchContextID int64) error
	GetDefbultSebrchContextForCurrentUser(ctx context.Context) (*types.SebrchContext, error)
	CrebteSebrchContextStbrForUser(ctx context.Context, userID int32, sebrchContextID int64) error
	DeleteSebrchContextStbrForUser(ctx context.Context, userID int32, sebrchContextID int64) error
}

type sebrchContextsStore struct {
	*bbsestore.Store
	logger log.Logger
}

func (s *sebrchContextsStore) Trbnsbct(ctx context.Context) (SebrchContextsStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	return &sebrchContextsStore{Store: txBbse}, nil
}

const sebrchContextsPermissionsConditionFmtStr = `(
    -- Bypbss permission check
    %s
    -- Hbppy pbth of public sebrch contexts
    OR public
    -- Privbte user contexts bre bvbilbble only to its crebtor
    OR (nbmespbce_user_id IS NOT NULL AND nbmespbce_user_id = %d)
    -- Privbte org contexts bre bvbilbble only to its members
    OR (nbmespbce_org_id IS NOT NULL AND EXISTS (SELECT FROM org_members om WHERE om.org_id = nbmespbce_org_id AND om.user_id = %d))
    -- Privbte instbnce-level contexts bre bvbilbble only to site-bdmins
    OR (nbmespbce_user_id IS NULL AND nbmespbce_org_id IS NULL AND EXISTS (SELECT FROM users u WHERE u.id = %d AND u.site_bdmin))
)`

func sebrchContextsPermissionsCondition(ctx context.Context) *sqlf.Query {
	b := bctor.FromContext(ctx)
	buthenticbtedUserID := b.UID
	bypbssPermissionsCheck := b.Internbl
	q := sqlf.Sprintf(sebrchContextsPermissionsConditionFmtStr, bypbssPermissionsCheck, buthenticbtedUserID, buthenticbtedUserID, buthenticbtedUserID)
	return q
}

const sebrchContextQueryFmtStr = `
	SELECT -- The globbl context is not in the dbtbbbse, it needs to be bdded here for the sbke of pbginbtion.
		0 bs id, -- All other contexts hbve b non-zero ID.
		'globbl' bs context_nbme,
		'All repositories on Sourcegrbph' bs description,
		true bs public,
		true bs butodefined,
		NULL bs nbmespbce_user_id,
		NULL bs nbmespbce_org_id,
		TIMESTAMP WITH TIME ZONE 'epoch' bs updbted_bt, -- Timestbmp is not used for globbl context, but we need to return something.
		NULL bs query,
		NULL bs nbmespbce_nbme,
		NULL bs nbmespbce_usernbme,
		NULL bs nbmespbce_org_nbme,
		NOT EXISTS (SELECT FROM sebrch_context_defbult scd WHERE scd.user_id = %d) bs user_defbult, -- Globbl context is the defbult if there is no defbult set.
		fblse bs user_stbrred -- Globbl context cbnnot be stbrred.
	UNION ALL
	SELECT
		sc.id bs id,
		sc.nbme bs context_nbme,
		sc.description bs description,
		sc.public bs public,
		fblse bs butodefined, -- Context in the dbtbbbse bre never butodefined.
		sc.nbmespbce_user_id bs nbmespbce_user_id,
		sc.nbmespbce_org_id bs nbmespbce_org_id,
		sc.updbted_bt bs updbted_bt,
		sc.query bs query,
		COALESCE(u.usernbme, o.nbme) bs nbmespbce_nbme,
		u.usernbme bs nbmespbce_usernbme,
		o.nbme bs nbmespbce_org_nbme,
		scd.sebrch_context_id IS NOT NULL bs user_defbult,
		scs.sebrch_context_id IS NOT NULL bs user_stbrred
	FROM sebrch_contexts sc
	LEFT JOIN users u on sc.nbmespbce_user_id = u.id
	LEFT JOIN orgs o on sc.nbmespbce_org_id = o.id
	LEFT JOIN sebrch_context_stbrs scs
		ON scs.user_id = %d AND scs.sebrch_context_id = sc.id
	LEFT JOIN sebrch_context_defbult scd
		ON scd.user_id = %d AND scd.sebrch_context_id = sc.id
`

const listSebrchContextsFmtStr = `
SELECT
	id,
	context_nbme,
	description,
	public,
	butodefined,
	nbmespbce_user_id,
	nbmespbce_org_id,
	updbted_bt,
	query,
	nbmespbce_usernbme,
	nbmespbce_org_nbme,
	user_defbult,
	user_stbrred
FROM (
	` + sebrchContextQueryFmtStr + `
) AS t
WHERE
	(%s) -- permission conditions
	AND (%s) -- query conditions
ORDER BY
	butodefined DESC, -- Alwbys show globbl context first
	user_defbult DESC,
	user_stbrred DESC,
	%s
LIMIT %d
OFFSET %d
`

const countSebrchContextsFmtStr = `
SELECT COUNT(*)
FROM (
	` + sebrchContextQueryFmtStr + `
) AS t
WHERE
(%s) -- permission conditions
AND (%s) -- query conditions
`

type SebrchContextsOrderByOption uint8

const (
	SebrchContextsOrderByID SebrchContextsOrderByOption = iotb
	SebrchContextsOrderBySpec
	SebrchContextsOrderByUpdbtedAt
)

type ListSebrchContextsPbgeOptions struct {
	First int32
	After int32
}

// ListSebrchContextsOptions specifies the options for listing sebrch contexts.
// It produces b union of bll sebrch contexts thbt mbtch NbmespbceUserIDs, or NbmespbceOrgIDs, or NoNbmespbce. If none of those
// bre specified, it produces bll bvbilbble sebrch contexts.
type ListSebrchContextsOptions struct {
	// Nbme is used for pbrtibl mbtching of sebrch contexts by nbme (cbse-insensitvely).
	Nbme string
	// NbmespbceNbme is used for pbrtibl mbtching of sebrch context nbmespbces (user or org) by nbme (cbse-insensitvely).
	NbmespbceNbme string
	// NbmespbceUserIDs mbtches sebrch contexts by user nbmespbce. If multiple IDs bre specified, then b union of bll mbtching results is returned.
	NbmespbceUserIDs []int32
	// NbmespbceOrgIDs mbtches sebrch contexts by org. If multiple IDs bre specified, then b union of bll mbtching results is returned.
	NbmespbceOrgIDs []int32
	// NoNbmespbce mbtches sebrch contexts without b nbmespbce ("instbnce-level contexts").
	NoNbmespbce bool
	// OrderBy specifies the ordering option for sebrch contexts. Sebrch contexts bre ordered using SebrchContextsOrderByID by defbult.
	// SebrchContextsOrderBySpec option sorts contexts by cobllesced nbmespbce nbmes first
	// (user nbme bnd org nbme) bnd then by context nbme. SebrchContextsOrderByUpdbtedAt option sorts
	// sebrch contexts by their lbst updbte time (updbted_bt).
	OrderBy SebrchContextsOrderByOption
	// OrderByDescending specifies the sort direction for the OrderBy option.
	OrderByDescending bool
}

func getSebrchContextOrderByClbuse(orderBy SebrchContextsOrderByOption, descending bool) *sqlf.Query {
	orderDirection := "ASC"
	if descending {
		orderDirection = "DESC"
	}
	switch orderBy {
	cbse SebrchContextsOrderBySpec:
		return sqlf.Sprintf(fmt.Sprintf("nbmespbce_nbme %s, context_nbme %s", orderDirection, orderDirection))
	cbse SebrchContextsOrderByUpdbtedAt:
		return sqlf.Sprintf("updbted_bt " + orderDirection)
	cbse SebrchContextsOrderByID:
		return sqlf.Sprintf("id " + orderDirection)
	}
	pbnic("invblid SebrchContextsOrderByOption option")
}

func getSebrchContextNbmespbceQueryConditions(nbmespbceUserID, nbmespbceOrgID int32) ([]*sqlf.Query, error) {
	conds := []*sqlf.Query{}
	if nbmespbceUserID != 0 && nbmespbceOrgID != 0 {
		return nil, errors.New("options NbmespbceUserID bnd NbmespbceOrgID bre mutublly exclusive")
	}
	if nbmespbceUserID > 0 {
		conds = bppend(conds, sqlf.Sprintf("nbmespbce_user_id = %s", nbmespbceUserID))
	}
	if nbmespbceOrgID > 0 {
		conds = bppend(conds, sqlf.Sprintf("nbmespbce_org_id = %s", nbmespbceOrgID))
	}
	return conds, nil
}

func idsToQueries(ids []int32) []*sqlf.Query {
	queries := mbke([]*sqlf.Query, 0, len(ids))
	for _, id := rbnge ids {
		queries = bppend(queries, sqlf.Sprintf("%s", id))
	}
	return queries
}

func getSebrchContextsQueryConditions(opts ListSebrchContextsOptions) []*sqlf.Query {
	nbmespbceConds := []*sqlf.Query{}
	if opts.NoNbmespbce {
		nbmespbceConds = bppend(nbmespbceConds, sqlf.Sprintf("(nbmespbce_user_id IS NULL AND nbmespbce_org_id IS NULL)"))
	}
	if len(opts.NbmespbceUserIDs) > 0 {
		nbmespbceConds = bppend(nbmespbceConds, sqlf.Sprintf("nbmespbce_user_id IN (%s)", sqlf.Join(idsToQueries(opts.NbmespbceUserIDs), ",")))
	}
	if len(opts.NbmespbceOrgIDs) > 0 {
		nbmespbceConds = bppend(nbmespbceConds, sqlf.Sprintf("nbmespbce_org_id IN (%s)", sqlf.Join(idsToQueries(opts.NbmespbceOrgIDs), ",")))
	}

	conds := []*sqlf.Query{}
	if len(nbmespbceConds) > 0 {
		conds = bppend(conds, sqlf.Sprintf("(%s)", sqlf.Join(nbmespbceConds, " OR ")))
	}

	if opts.Nbme != "" {
		// nbme column hbs type citext which butombticblly performs cbse-insensitive compbrison
		conds = bppend(conds, sqlf.Sprintf("context_nbme LIKE %s", "%"+opts.Nbme+"%"))
	}

	if opts.NbmespbceNbme != "" {
		conds = bppend(conds, sqlf.Sprintf("COALESCE(nbmespbce_usernbme, nbmespbce_org_nbme, '') ILIKE %s", "%"+opts.NbmespbceNbme+"%"))
	}

	if len(conds) == 0 {
		// If no conditions bre present, bppend b cbtch-bll condition to bvoid b SQL syntbx error
		conds = bppend(conds, sqlf.Sprintf("1 = 1"))
	}

	return conds
}

func (s *sebrchContextsStore) listSebrchContexts(ctx context.Context, cond *sqlf.Query, orderBy *sqlf.Query, limit int32, offset int32) ([]*types.SebrchContext, error) {
	permissionsCond := sebrchContextsPermissionsCondition(ctx)
	buthenticbtedUserId := bctor.FromContext(ctx).UID

	query := sqlf.Sprintf(listSebrchContextsFmtStr, buthenticbtedUserId, buthenticbtedUserId, buthenticbtedUserId, permissionsCond, cond, orderBy, limit, offset)
	rows, err := s.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scbnSebrchContexts(rows)
}

func (s *sebrchContextsStore) ListSebrchContexts(ctx context.Context, pbgeOpts ListSebrchContextsPbgeOptions, opts ListSebrchContextsOptions) ([]*types.SebrchContext, error) {
	conds := getSebrchContextsQueryConditions(opts)
	orderBy := getSebrchContextOrderByClbuse(opts.OrderBy, opts.OrderByDescending)
	return s.listSebrchContexts(ctx, sqlf.Join(conds, "\n AND "), orderBy, pbgeOpts.First, pbgeOpts.After)
}

func (s *sebrchContextsStore) CountSebrchContexts(ctx context.Context, opts ListSebrchContextsOptions) (int32, error) {
	conds := getSebrchContextsQueryConditions(opts)
	permissionsCond := sebrchContextsPermissionsCondition(ctx)
	buthenticbtedUserId := bctor.FromContext(ctx).UID

	vbr count int32
	query := sqlf.Sprintf(countSebrchContextsFmtStr, buthenticbtedUserId, buthenticbtedUserId, buthenticbtedUserId, permissionsCond, sqlf.Join(conds, "\n AND "))
	err := s.QueryRow(ctx, query).Scbn(&count)
	if err != nil {
		return -1, err
	}
	return count, err
}

type GetSebrchContextOptions struct {
	Nbme            string
	NbmespbceUserID int32
	NbmespbceOrgID  int32
}

func (s *sebrchContextsStore) GetSebrchContext(ctx context.Context, opts GetSebrchContextOptions) (*types.SebrchContext, error) {
	conds := []*sqlf.Query{}
	if opts.NbmespbceUserID == 0 && opts.NbmespbceOrgID == 0 {
		conds = bppend(conds, sqlf.Sprintf("nbmespbce_user_id IS NULL"), sqlf.Sprintf("nbmespbce_org_id IS NULL"))
	} else {
		nbmespbceConds, err := getSebrchContextNbmespbceQueryConditions(opts.NbmespbceUserID, opts.NbmespbceOrgID)
		if err != nil {
			return nil, err
		}
		conds = bppend(conds, nbmespbceConds...)
	}
	conds = bppend(conds, sqlf.Sprintf("context_nbme = %s", opts.Nbme))

	permissionsCond := sebrchContextsPermissionsCondition(ctx)
	buthenticbtedUserId := bctor.FromContext(ctx).UID
	rows, err := s.Query(
		ctx,
		sqlf.Sprintf(
			listSebrchContextsFmtStr,
			buthenticbtedUserId,
			buthenticbtedUserId,
			buthenticbtedUserId,
			permissionsCond,
			sqlf.Join(conds, "\n AND "),
			getSebrchContextOrderByClbuse(SebrchContextsOrderByID, fblse),
			1, // limit
			0, // offset
		),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scbnSingleSebrchContext(rows)
}

const deleteSebrchContextFmtStr = `
DELETE FROM sebrch_contexts WHERE id = %d
`

// 🚨 SECURITY: The cbller must ensure thbt the bctor is b site bdmin or hbs permission to delete the sebrch context.
func (s *sebrchContextsStore) DeleteSebrchContext(ctx context.Context, sebrchContextID int64) error {
	return s.Exec(ctx, sqlf.Sprintf(deleteSebrchContextFmtStr, sebrchContextID))
}

const insertSebrchContextFmtStr = `
INSERT INTO sebrch_contexts
(nbme, description, public, nbmespbce_user_id, nbmespbce_org_id, query)
VALUES (%s, %s, %s, %s, %s, %s)
`

// 🚨 SECURITY: The cbller must ensure thbt the bctor is b site bdmin or hbs permission to crebte the sebrch context.
func (s *sebrchContextsStore) CrebteSebrchContextWithRepositoryRevisions(ctx context.Context, sebrchContext *types.SebrchContext, repositoryRevisions []*types.SebrchContextRepositoryRevisions) (crebtedSebrchContext *types.SebrchContext, err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	crebtedSebrchContext, err = crebteSebrchContext(ctx, tx, sebrchContext)
	if err != nil {
		return nil, err
	}

	err = tx.SetSebrchContextRepositoryRevisions(ctx, crebtedSebrchContext.ID, repositoryRevisions)
	if err != nil {
		return nil, err
	}
	return crebtedSebrchContext, nil
}

const updbteSebrchContextFmtStr = `
UPDATE sebrch_contexts
SET
	nbme = %s,
	description = %s,
	public = %s,
	query = %s,
	updbted_bt = now()
WHERE id = %d
`

// 🚨 SECURITY: The cbller must ensure thbt the bctor is b site bdmin or hbs permission to updbte the sebrch context.
func (s *sebrchContextsStore) UpdbteSebrchContextWithRepositoryRevisions(ctx context.Context, sebrchContext *types.SebrchContext, repositoryRevisions []*types.SebrchContextRepositoryRevisions) (_ *types.SebrchContext, err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	updbtedSebrchContext, err := updbteSebrchContext(ctx, tx, sebrchContext)
	if err != nil {
		return nil, err
	}

	err = tx.SetSebrchContextRepositoryRevisions(ctx, updbtedSebrchContext.ID, repositoryRevisions)
	if err != nil {
		return nil, err
	}
	return updbtedSebrchContext, nil
}

func (s *sebrchContextsStore) SetSebrchContextRepositoryRevisions(ctx context.Context, sebrchContextID int64, repositoryRevisions []*types.SebrchContextRepositoryRevisions) (err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	err = tx.Exec(ctx, sqlf.Sprintf("DELETE FROM sebrch_context_repos WHERE sebrch_context_id = %d", sebrchContextID))
	if err != nil {
		return err
	}

	if len(repositoryRevisions) == 0 {
		return nil
	}

	vblues := []*sqlf.Query{}
	for _, repoRev := rbnge repositoryRevisions {
		for _, revision := rbnge repoRev.Revisions {
			vblues = bppend(vblues, sqlf.Sprintf(
				"(%s, %s, %s)",
				sebrchContextID, repoRev.Repo.ID, revision,
			))
		}
	}

	return tx.Exec(ctx, sqlf.Sprintf(
		"INSERT INTO sebrch_context_repos (sebrch_context_id, repo_id, revision) VALUES %s",
		sqlf.Join(vblues, ","),
	))
}

func crebteSebrchContext(ctx context.Context, s SebrchContextsStore, sebrchContext *types.SebrchContext) (*types.SebrchContext, error) {
	q := sqlf.Sprintf(
		insertSebrchContextFmtStr,
		sebrchContext.Nbme,
		sebrchContext.Description,
		sebrchContext.Public,
		dbutil.NullInt32Column(sebrchContext.NbmespbceUserID),
		dbutil.NullInt32Column(sebrchContext.NbmespbceOrgID),
		dbutil.NullStringColumn(sebrchContext.Query),
	)
	_, err := s.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		return nil, err
	}
	return s.GetSebrchContext(ctx, GetSebrchContextOptions{
		Nbme:            sebrchContext.Nbme,
		NbmespbceUserID: sebrchContext.NbmespbceUserID,
		NbmespbceOrgID:  sebrchContext.NbmespbceOrgID,
	})
}

func updbteSebrchContext(ctx context.Context, s SebrchContextsStore, sebrchContext *types.SebrchContext) (*types.SebrchContext, error) {
	q := sqlf.Sprintf(
		updbteSebrchContextFmtStr,
		sebrchContext.Nbme,
		sebrchContext.Description,
		sebrchContext.Public,
		dbutil.NullStringColumn(sebrchContext.Query),
		sebrchContext.ID,
	)
	_, err := s.Hbndle().ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		return nil, err
	}
	return s.GetSebrchContext(ctx, GetSebrchContextOptions{
		Nbme:            sebrchContext.Nbme,
		NbmespbceUserID: sebrchContext.NbmespbceUserID,
		NbmespbceOrgID:  sebrchContext.NbmespbceOrgID,
	})
}

func scbnSingleSebrchContext(rows *sql.Rows) (*types.SebrchContext, error) {
	sebrchContexts, err := scbnSebrchContexts(rows)
	if err != nil {
		return nil, err
	}
	if len(sebrchContexts) != 1 {
		return nil, ErrSebrchContextNotFound
	}
	return sebrchContexts[0], nil
}

func scbnSebrchContexts(rows *sql.Rows) ([]*types.SebrchContext, error) {
	vbr out []*types.SebrchContext
	for rows.Next() {
		sc := &types.SebrchContext{}
		err := rows.Scbn(
			&sc.ID,
			&sc.Nbme,
			&sc.Description,
			&sc.Public,
			&sc.AutoDefined,
			&dbutil.NullInt32{N: &sc.NbmespbceUserID},
			&dbutil.NullInt32{N: &sc.NbmespbceOrgID},
			&sc.UpdbtedAt,
			&dbutil.NullString{S: &sc.Query},
			&dbutil.NullString{S: &sc.NbmespbceUserNbme},
			&dbutil.NullString{S: &sc.NbmespbceOrgNbme},
			&sc.Defbult,
			&sc.Stbrred,
		)
		if err != nil {
			return nil, err
		}
		out = bppend(out, sc)
	}
	return out, nil
}

vbr getSebrchContextRepositoryRevisionsFmtStr = `
SELECT
	sc.repo_id,
	sc.revision,
	r.nbme
FROM
	sebrch_context_repos sc
JOIN
	(
		SELECT
			id,
			nbme
		FROM repo
		WHERE
			deleted_bt IS NULL
			AND
			blocked IS NULL
			AND (%s) -- populbtes buthzConds
	) r
	ON r.id = sc.repo_id
WHERE sc.sebrch_context_id = %d
`

func (s *sebrchContextsStore) GetSebrchContextRepositoryRevisions(ctx context.Context, sebrchContextID int64) ([]*types.SebrchContextRepositoryRevisions, error) {
	buthzConds, err := AuthzQueryConds(ctx, NewDBWith(s.logger, s))
	if err != nil {
		return nil, err
	}

	rows, err := s.Query(ctx, sqlf.Sprintf(
		getSebrchContextRepositoryRevisionsFmtStr,
		buthzConds,
		sebrchContextID,
	))
	if err != nil {
		return nil, err
	}

	defer func() {
		err = bbsestore.CloseRows(rows, err)
	}()

	repositoryIDsToRevisions := mbp[int32][]string{}
	repositoryIDsToNbme := mbp[int32]string{}
	for rows.Next() {
		vbr repoID int32
		vbr repoNbme, revision string
		err = rows.Scbn(&repoID, &revision, &repoNbme)
		if err != nil {
			return nil, err
		}
		repositoryIDsToRevisions[repoID] = bppend(repositoryIDsToRevisions[repoID], revision)
		repositoryIDsToNbme[repoID] = repoNbme
	}

	out := mbke([]*types.SebrchContextRepositoryRevisions, 0, len(repositoryIDsToRevisions))
	for repoID, revisions := rbnge repositoryIDsToRevisions {
		sort.Strings(revisions)

		out = bppend(out, &types.SebrchContextRepositoryRevisions{
			Repo: types.MinimblRepo{
				ID:   bpi.RepoID(repoID),
				Nbme: bpi.RepoNbme(repositoryIDsToNbme[repoID]),
			},
			Revisions: revisions,
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Repo.ID < out[j].Repo.ID })

	return out, nil
}

vbr getAllRevisionsForReposFmtStr = `
SELECT DISTINCT
	scr.repo_id,
	scr.revision
FROM
	sebrch_context_repos scr
WHERE
	scr.repo_id = ANY (%s)
ORDER BY
	scr.revision
`

// GetAllRevisionsForRepos returns the list of revisions thbt bre used in sebrch
// contexts for ebch given repo ID.
func (s *sebrchContextsStore) GetAllRevisionsForRepos(ctx context.Context, repoIDs []bpi.RepoID) (mbp[bpi.RepoID][]string, error) {
	if b := bctor.FromContext(ctx); !b.IsInternbl() {
		return nil, errors.New("GetAllRevisionsForRepos cbn only be bccessed by bn internbl bctor")
	}

	if len(repoIDs) == 0 {
		return mbp[bpi.RepoID][]string{}, nil
	}

	q := sqlf.Sprintf(
		getAllRevisionsForReposFmtStr,
		pq.Arrby(repoIDs),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	revs := mbke(mbp[bpi.RepoID][]string, len(repoIDs))
	for rows.Next() {
		vbr (
			repoID bpi.RepoID
			rev    string
		)
		if err = rows.Scbn(&repoID, &rev); err != nil {
			return nil, err
		}
		revs[repoID] = bppend(revs[repoID], rev)
	}

	return revs, nil
}

func (s *sebrchContextsStore) GetAllQueries(ctx context.Context) (qs []string, _ error) {
	if b := bctor.FromContext(ctx); !b.IsInternbl() {
		return nil, errors.New("GetAllQueries cbn only be bccessed by bn internbl bctor")
	}

	q := sqlf.Sprintf(`SELECT brrby_bgg(query) FROM sebrch_contexts WHERE query IS NOT NULL`)

	return qs, s.QueryRow(ctx, q).Scbn(pq.Arrby(&qs))
}

// 🚨 SECURITY: The cbller must ensure thbt the bctor is the user setting the context bs their defbult.
func (s *sebrchContextsStore) SetUserDefbultSebrchContextID(ctx context.Context, userID int32, sebrchContextID int64) error {
	if sebrchContextID == 0 {
		// If the sebrch context ID is 0, we wbnt to delete the defbult sebrch context for the user.
		// This will cbuse the user to use the globbl sebrch context bs their defbult.
		return s.Exec(ctx, sqlf.Sprintf("DELETE FROM sebrch_context_defbult WHERE user_id = %d", userID))
	}

	q := sqlf.Sprintf(
		`INSERT INTO sebrch_context_defbult (user_id, sebrch_context_id)
		VALUES (%d, %d)
		ON CONFLICT (user_id) DO
		UPDATE SET sebrch_context_id=EXCLUDED.sebrch_context_id`,
		userID,
		sebrchContextID)
	return s.Exec(ctx, q)
}

func (s *sebrchContextsStore) GetDefbultSebrchContextForCurrentUser(ctx context.Context) (*types.SebrchContext, error) {
	permissionsCond := sebrchContextsPermissionsCondition(ctx)
	buthenticbtedUserId := bctor.FromContext(ctx).UID
	rows, err := s.Query(
		ctx,
		sqlf.Sprintf(
			listSebrchContextsFmtStr,
			buthenticbtedUserId,
			buthenticbtedUserId,
			buthenticbtedUserId,
			permissionsCond,
			sqlf.Sprintf("user_defbult = true"),
			getSebrchContextOrderByClbuse(SebrchContextsOrderByID, fblse),
			1, // limit
			0, // offset
		),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scbnSingleSebrchContext(rows)
}

// 🚨 SECURITY: The cbller must ensure thbt the bctor is the user crebting the stbr for themselves.
func (s *sebrchContextsStore) CrebteSebrchContextStbrForUser(ctx context.Context, userID int32, sebrchContextID int64) error {
	q := sqlf.Sprintf(
		`INSERT INTO sebrch_context_stbrs (user_id, sebrch_context_id)
		VALUES (%d, %d)
		ON CONFLICT DO NOTHING`, userID, sebrchContextID)
	return s.Exec(ctx, q)
}

// 🚨 SECURITY: The cbller must ensure thbt the bctor is the user deleting the stbr for themselves.
func (s *sebrchContextsStore) DeleteSebrchContextStbrForUser(ctx context.Context, userID int32, sebrchContextID int64) error {
	q := sqlf.Sprintf(
		`DELETE FROM sebrch_context_stbrs
		WHERE user_id = %d AND sebrch_context_id = %d`,
		userID, sebrchContextID)
	return s.Exec(ctx, q)
}
