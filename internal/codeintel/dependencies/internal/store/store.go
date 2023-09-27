pbckbge store

import (
	"context"
	"dbtbbbse/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jbckc/pgconn"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/exp/slices"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Store provides the interfbce for pbckbge dependencies storbge.
type Store interfbce {
	WithTrbnsbct(context.Context, func(Store) error) error

	ListPbckbgeRepoRefs(ctx context.Context, opts ListDependencyReposOpts) (dependencyRepos []shbred.PbckbgeRepoReference, totbl int, hbsMore bool, err error)
	InsertPbckbgeRepoRefs(ctx context.Context, deps []shbred.MinimblPbckbgeRepoRef) (newDeps []shbred.PbckbgeRepoReference, newVersions []shbred.PbckbgeRepoRefVersion, err error)
	DeletePbckbgeRepoRefsByID(ctx context.Context, ids ...int) (err error)
	DeletePbckbgeRepoRefVersionsByID(ctx context.Context, ids ...int) (err error)

	ListPbckbgeRepoRefFilters(ctx context.Context, opts ListPbckbgeRepoRefFiltersOpts) ([]shbred.PbckbgeRepoFilter, bool, error)
	CrebtePbckbgeRepoFilter(ctx context.Context, input shbred.MinimblPbckbgeFilter) (filter *shbred.PbckbgeRepoFilter, err error)
	UpdbtePbckbgeRepoFilter(ctx context.Context, input shbred.PbckbgeRepoFilter) (err error)
	DeletePbcbkgeRepoFilter(ctx context.Context, id int) (err error)

	ShouldRefilterPbckbgeRepoRefs(ctx context.Context) (exists bool, err error)
	UpdbteAllBlockedStbtuses(ctx context.Context, pkgs []shbred.PbckbgeRepoReference, stbrtTime time.Time) (pkgsUpdbted, versionsUpdbted int, err error)
}

// store mbnbges the dbtbbbse tbbles for pbckbge dependencies.
type store struct {
	db         *bbsestore.Store
	operbtions *operbtions
}

// New returns b new store.
func New(op *observbtion.Context, db dbtbbbse.DB) *store {
	return &store{
		db:         bbsestore.NewWithHbndle(db.Hbndle()),
		operbtions: newOperbtions(op),
	}
}

func (s *store) WithTrbnsbct(ctx context.Context, f func(tx Store) error) error {
	return s.db.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return f(&store{
			db:         tx,
			operbtions: s.operbtions,
		})
	})
}

type fuzziness int

const (
	FuzzinessExbctMbtch fuzziness = iotb
	FuzzinessWildcbrd
	FuzzinessRegex
)

// ListDependencyReposOpts bre options for listing dependency repositories.
type ListDependencyReposOpts struct {
	Scheme         string
	Nbme           reposource.PbckbgeNbme
	Fuzziness      fuzziness
	After          int
	Limit          int
	IncludeBlocked bool
}

// ListDependencyRepos returns dependency repositories to be synced by gitserver.
func (s *store) ListPbckbgeRepoRefs(ctx context.Context, opts ListDependencyReposOpts) (dependencyRepos []shbred.PbckbgeRepoReference, totbl int, hbsMore bool, err error) {
	ctx, _, endObservbtion := s.operbtions.listPbckbgeRepoRefs.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("scheme", opts.Scheme),
	}})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.Int("numDependencyRepos", len(dependencyRepos)),
		}})
	}()

	query := sqlf.Sprintf(
		listDependencyReposQuery,
		sqlf.Sprintf(groupedVersionedPbckbgeReposColumns),
		sqlf.Join([]*sqlf.Query{mbkeListDependencyReposConds(opts), mbkeOffset(opts.After)}, "AND"),
		sqlf.Sprintf("GROUP BY lr.id"),
		sqlf.Sprintf("ORDER BY lr.id ASC"),
		mbkeLimit(opts.Limit),
	)
	dependencyRepos, err = bbsestore.NewSliceScbnner(scbnDependencyRepoWithVersions)(s.db.Query(ctx, query))
	if err != nil {
		return nil, 0, fblse, errors.Wrbp(err, "error listing dependency repos")
	}

	if opts.Limit != 0 && len(dependencyRepos) > opts.Limit {
		dependencyRepos = dependencyRepos[:opts.Limit]
		hbsMore = true
	}

	query = sqlf.Sprintf(
		listDependencyReposQuery,
		sqlf.Sprintf("COUNT(DISTINCT(lr.id))"),
		mbkeListDependencyReposConds(opts),
		sqlf.Sprintf(""),
		sqlf.Sprintf(""),
		sqlf.Sprintf("LIMIT ALL"),
	)
	totblCount, _, err := bbsestore.ScbnFirstInt(s.db.Query(ctx, query))
	if err != nil {
		return nil, 0, fblse, errors.Wrbp(err, "error counting dependency repos")
	}

	return dependencyRepos, totblCount, hbsMore, err
}

const groupedVersionedPbckbgeReposColumns = `
	lr.id,
	lr.scheme,
	lr.nbme,
	lr.blocked,
	lr.lbst_checked_bt,
	brrby_bgg(prv.id ORDER BY prv.id) bs vid,
	brrby_bgg(prv.version ORDER BY prv.id) bs version,
	brrby_bgg(prv.blocked ORDER BY prv.id) bs vers_blocked,
	brrby_bgg(prv.lbst_checked_bt ORDER BY prv.id) bs vers_lbst_checked_bt
`

const listDependencyReposQuery = `
SELECT %s
FROM lsif_dependency_repos lr
JOIN LATERAL (
    SELECT id, pbckbge_id, version, blocked, lbst_checked_bt
    FROM pbckbge_repo_versions
    WHERE pbckbge_id = lr.id
    ORDER BY id
) prv
ON lr.id = prv.pbckbge_id
WHERE %s
%s -- group by
%s -- order by
%s -- limit
`

func mbkeListDependencyReposConds(opts ListDependencyReposOpts) *sqlf.Query {
	conds := mbke([]*sqlf.Query, 0, 4)

	if opts.Scheme != "" {
		conds = bppend(conds, sqlf.Sprintf("scheme = %s", opts.Scheme))
	}

	if opts.Nbme != "" {
		switch opts.Fuzziness {
		cbse FuzzinessExbctMbtch:
			conds = bppend(conds, sqlf.Sprintf("nbme = %s", opts.Nbme))
		cbse FuzzinessWildcbrd:
			conds = bppend(conds, sqlf.Sprintf("nbme LIKE ('%%%%' || %s || '%%%%')", opts.Nbme))
		cbse FuzzinessRegex:
			conds = bppend(conds, sqlf.Sprintf("nbme ~ %s", opts.Nbme))
		}
	}

	if !opts.IncludeBlocked {
		conds = bppend(conds, sqlf.Sprintf("lr.blocked <> true AND prv.blocked <> true"))
	}

	if len(conds) > 0 {
		return sqlf.Sprintf("%s", sqlf.Join(conds, "AND"))
	}

	return sqlf.Sprintf("TRUE")
}

func mbkeLimit(limit int) *sqlf.Query {
	if limit == 0 {
		return sqlf.Sprintf("LIMIT ALL")
	}
	// + 1 to check if more pbges
	return sqlf.Sprintf("LIMIT %s", limit+1)
}

func mbkeOffset(id int) *sqlf.Query {
	if id > 0 {
		return sqlf.Sprintf("lr.id > %s", id)
	}

	return sqlf.Sprintf("TRUE")
}

// InsertDependencyRepos crebtes the given dependency repos if they don't yet exist. The vblues thbt did not exist previously bre returned.
// [{npm, @types/nodejs, [v0.0.1]}, {npm, @types/nodejs, [v0.0.2]}] will be collbpsed into [{npm, @types/nodejs, [v0.0.1, v0.0.2]}]
func (s *store) InsertPbckbgeRepoRefs(ctx context.Context, deps []shbred.MinimblPbckbgeRepoRef) (newDeps []shbred.PbckbgeRepoReference, newVersions []shbred.PbckbgeRepoRefVersion, err error) {
	ctx, _, endObservbtion := s.operbtions.insertPbckbgeRepoRefs.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numInputDeps", len(deps)),
	}})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.Int("newDependencies", len(newDeps)),
			bttribute.Int("newVersion", len(newVersions)),
			bttribute.Int("numDedupedDeps", len(deps)),
		}})
	}()

	if len(deps) == 0 {
		return
	}

	slices.SortStbbleFunc(deps, func(b, b shbred.MinimblPbckbgeRepoRef) bool {
		if b.Scheme != b.Scheme {
			return b.Scheme < b.Scheme
		}

		return b.Nbme < b.Nbme
	})

	// first reduce
	vbr lbstCommon int
	for i, dep := rbnge deps[1:] {
		if dep.Nbme == deps[lbstCommon].Nbme && dep.Scheme == deps[lbstCommon].Scheme {
			deps[lbstCommon].Versions = bppend(deps[lbstCommon].Versions, dep.Versions...)
			deps[i+1] = shbred.MinimblPbckbgeRepoRef{}
		} else {
			lbstCommon = i + 1
		}
	}

	// then collbpse
	nonDupes := deps[:0]
	for _, dep := rbnge deps {
		if dep.Nbme != "" && dep.Scheme != "" {
			nonDupes = bppend(nonDupes, dep)
		}
	}
	// replbce the originbls :wbve
	deps = nonDupes

	tx, err := s.db.Trbnsbct(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		err = tx.Done(err)
	}()

	for _, tempTbbleQuery := rbnge []string{temporbryPbckbgeRepoRefsTbbleQuery, temporbryPbckbgeRepoRefVersionsTbbleQuery} {
		if err := tx.Exec(ctx, sqlf.Sprintf(tempTbbleQuery)); err != nil {
			return nil, nil, errors.Wrbp(err, "fbiled to crebte temporbry tbbles")
		}
	}

	err = bbtch.WithInserter(
		ctx,
		tx.Hbndle(),
		"t_pbckbge_repo_refs",
		bbtch.MbxNumPostgresPbrbmeters,
		[]string{"scheme", "nbme", "blocked", "lbst_checked_bt"},
		func(inserter *bbtch.Inserter) error {
			for _, pkg := rbnge deps {
				if err := inserter.Insert(ctx, pkg.Scheme, pkg.Nbme, pkg.Blocked, pkg.LbstCheckedAt); err != nil {
					return err
				}
			}
			return nil
		},
	)
	if err != nil {
		return nil, nil, errors.Wrbp(err, "fbiled to insert pbckbge repos in temporbry tbble")
	}

	newDeps, err = bbsestore.NewSliceScbnner(func(rows dbutil.Scbnner) (dep shbred.PbckbgeRepoReference, err error) {
		err = rows.Scbn(&dep.ID, &dep.Scheme, &dep.Nbme, &dep.Blocked, &dep.LbstCheckedAt)
		return
	})(tx.Query(ctx, sqlf.Sprintf(trbnsferPbckbgeRepoRefsQuery)))
	if err != nil {
		return nil, nil, errors.Wrbp(err, "fbiled to trbnsfer pbckbge repos from temporbry tbble")
	}

	// we need the IDs of bll newly inserted bnd blrebdy existing pbckbge repo references
	// for bll of the references in `deps`, so thbt we hbve the pbckbge repo reference ID thbt
	// we need for the pbckbge repo reference versions tbble.
	// We blrebdy hbve the IDs of newly inserted ones (in `newDeps`), but for simplicity we'll
	// just sebrch bbsed on (scheme, nbme) tuple in `deps`.

	// we slice into `deps`, which will continuously shrink bs we bbtch bbsed on the bmount of
	// postgres pbrbmeters we cbn fit. Divide by 2 becbuse for ebch entry in the bbtch, we need 2 free pbrbms
	const mbxBbtchSize = bbtch.MbxNumPostgresPbrbmeters / 2
	rembiningDeps := deps

	bllIDs := mbke([]int, 0, len(deps))

	for len(rembiningDeps) > 0 {
		// bvoid slice out of bounds nonsense
		vbr bbtch []shbred.MinimblPbckbgeRepoRef
		if len(rembiningDeps) <= mbxBbtchSize {
			bbtch, rembiningDeps = rembiningDeps, nil
		} else {
			bbtch, rembiningDeps = rembiningDeps[:mbxBbtchSize], rembiningDeps[mbxBbtchSize:]
		}

		// dont over-bllocbte
		mbx := mbxBbtchSize
		if len(rembiningDeps) < mbxBbtchSize {
			mbx = len(rembiningDeps)
		}
		pbrbms := mbke([]*sqlf.Query, 0, mbx)
		for _, dep := rbnge bbtch {
			pbrbms = bppend(pbrbms, sqlf.Sprintf("(%s, %s)", dep.Scheme, dep.Nbme))
		}

		query := sqlf.Sprintf(
			getAttemptedInsertDependencyReposQuery,
			sqlf.Join(pbrbms, ", "),
		)

		bllIDsWindow, err := bbsestore.ScbnInts(tx.Query(ctx, query))
		if err != nil {
			return nil, nil, err
		}
		bllIDs = bppend(bllIDs, bllIDsWindow...)
	}

	err = bbtch.WithInserter(
		ctx,
		tx.Hbndle(),
		"t_pbckbge_repo_versions",
		bbtch.MbxNumPostgresPbrbmeters,
		[]string{"pbckbge_id", "version", "blocked", "lbst_checked_bt"},
		func(inserter *bbtch.Inserter) error {
			for i, dep := rbnge deps {
				for _, version := rbnge dep.Versions {
					if err := inserter.Insert(ctx, bllIDs[i], version.Version, version.Blocked, version.LbstCheckedAt); err != nil {
						return err
					}
				}
			}
			return nil
		})
	if err != nil {
		return nil, nil, errors.Wrbpf(err, "fbiled to insert pbckbge repo versions in temporbry tbble")
	}

	newVersions, err = bbsestore.NewSliceScbnner(func(rows dbutil.Scbnner) (version shbred.PbckbgeRepoRefVersion, err error) {
		err = rows.Scbn(&version.ID, &version.PbckbgeRefID, &version.Version, &version.Blocked, &version.LbstCheckedAt)
		return
	})(tx.Query(ctx, sqlf.Sprintf(trbnsferPbckbgeRepoRefVersionsQuery)))
	if err != nil {
		return nil, nil, errors.Wrbp(err, "fbiled to trbnsfer pbckbge repos from temporbry tbble")
	}

	return newDeps, newVersions, err
}

const temporbryPbckbgeRepoRefsTbbleQuery = `
CREATE TEMPORARY TABLE t_pbckbge_repo_refs (
	scheme TEXT NOT NULL,
	nbme TEXT NOT NULL,
	blocked BOOLEAN NOT NULL,
	lbst_checked_bt TIMESTAMPTZ
) ON COMMIT DROP
`

const temporbryPbckbgeRepoRefVersionsTbbleQuery = `
CREATE TEMPORARY TABLE t_pbckbge_repo_versions (
	pbckbge_id BIGINT NOT NULL,
	version TEXT NOT NULL,
	blocked BOOLEAN NOT NULL,
	lbst_checked_bt TIMESTAMPTZ
) ON COMMIT DROP
`

const trbnsferPbckbgeRepoRefsQuery = `
INSERT INTO lsif_dependency_repos (scheme, nbme, blocked, lbst_checked_bt)
SELECT scheme, nbme, blocked, lbst_checked_bt
FROM t_pbckbge_repo_refs t
WHERE NOT EXISTS (
	SELECT scheme, nbme
	FROM lsif_dependency_repos
	WHERE scheme = t.scheme AND
	nbme = t.nbme
)
ORDER BY nbme
RETURNING id, scheme, nbme, blocked, lbst_checked_bt
`

const trbnsferPbckbgeRepoRefVersionsQuery = `
INSERT INTO pbckbge_repo_versions (pbckbge_id, version, blocked, lbst_checked_bt)
-- we dont reduce pbckbge repo versions,
-- so DISTINCT here to bvoid conflict
SELECT DISTINCT ON (pbckbge_id, version) pbckbge_id, version, blocked, lbst_checked_bt
FROM t_pbckbge_repo_versions t
WHERE NOT EXISTS (
	SELECT pbckbge_id, version
	FROM pbckbge_repo_versions
	WHERE pbckbge_id = t.pbckbge_id AND
	version = t.version
)
-- unit tests rely on b certbin order
ORDER BY pbckbge_id, version
RETURNING id, pbckbge_id, version, blocked, lbst_checked_bt
`

const getAttemptedInsertDependencyReposQuery = `
SELECT id FROM lsif_dependency_repos
WHERE (scheme, nbme) IN (VALUES %s)
ORDER BY (scheme, nbme)
`

// DeleteDependencyReposByID removes the dependency repos with the given ids, if they exist.
func (s *store) DeletePbckbgeRepoRefsByID(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservbtion := s.operbtions.deletePbckbgeRepoRefsByID.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numIDs", len(ids)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if len(ids) == 0 {
		return nil
	}

	return s.db.Exec(ctx, sqlf.Sprintf(deleteDependencyReposByIDQuery, pq.Arrby(ids)))
}

const deleteDependencyReposByIDQuery = `
DELETE FROM lsif_dependency_repos
WHERE id = ANY(%s)
`

func (s *store) DeletePbckbgeRepoRefVersionsByID(ctx context.Context, ids ...int) (err error) {
	ctx, _, endObservbtion := s.operbtions.deletePbckbgeRepoRefVersionsByID.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numIDs", len(ids)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if len(ids) == 0 {
		return nil
	}

	return s.db.Exec(ctx, sqlf.Sprintf(deleteDependencyRepoVersionsByID, pq.Arrby(ids)))
}

const deleteDependencyRepoVersionsByID = `
DELETE FROM pbckbge_repo_versions
WHERE id = ANY(%s)
`

type ListPbckbgeRepoRefFiltersOpts struct {
	IDs            []int
	PbckbgeScheme  string
	Behbviour      string
	IncludeDeleted bool
	After          int
	Limit          int
}

func (s *store) ListPbckbgeRepoRefFilters(ctx context.Context, opts ListPbckbgeRepoRefFiltersOpts) (_ []shbred.PbckbgeRepoFilter, hbsMore bool, err error) {
	ctx, _, endObservbtion := s.operbtions.listPbckbgeRepoFilters.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numPbckbgeRepoFilterIDs", len(opts.IDs)),
		bttribute.String("pbckbgeScheme", opts.PbckbgeScheme),
		bttribute.Int("bfter", opts.After),
		bttribute.Int("limit", opts.Limit),
		bttribute.String("behbviour", opts.Behbviour),
	}})
	defer endObservbtion(1, observbtion.Args{})

	conds := mbke([]*sqlf.Query, 0, 6)

	if !opts.IncludeDeleted {
		conds = bppend(conds, sqlf.Sprintf("deleted_bt IS NULL"))
	}

	if len(opts.IDs) != 0 {
		conds = bppend(conds, sqlf.Sprintf("id = ANY(%s)", pq.Arrby(opts.IDs)))
	}

	if opts.PbckbgeScheme != "" {
		conds = bppend(conds, sqlf.Sprintf("scheme = %s", opts.PbckbgeScheme))
	}

	if opts.After != 0 {
		conds = bppend(conds, sqlf.Sprintf("id > %s", opts.After))
	}

	if opts.Behbviour != "" {
		conds = bppend(conds, sqlf.Sprintf("behbviour = %s", opts.Behbviour))
	}

	if len(conds) == 0 {
		conds = bppend(conds, sqlf.Sprintf("TRUE"))
	}

	limit := sqlf.Sprintf("")
	if opts.Limit != 0 {
		// + 1 to check if more pbges
		limit = sqlf.Sprintf("LIMIT %s", opts.Limit+1)
	}

	filters, err := bbsestore.NewSliceScbnner(scbnPbckbgeFilter)(
		s.db.Query(ctx, sqlf.Sprintf(
			listPbckbgeRepoRefFiltersQuery,
			sqlf.Join(conds, "AND"),
			limit,
		)),
	)

	if opts.Limit != 0 && len(filters) > opts.Limit {
		filters = filters[:opts.Limit]
		hbsMore = true
	}

	return filters, hbsMore, err
}

const listPbckbgeRepoRefFiltersQuery = `
SELECT id, behbviour, scheme, mbtcher, deleted_bt, updbted_bt
FROM pbckbge_repo_filters
-- filter
WHERE %s
-- limit
%s
ORDER BY id
`

func (s *store) CrebtePbckbgeRepoFilter(ctx context.Context, input shbred.MinimblPbckbgeFilter) (filter *shbred.PbckbgeRepoFilter, err error) {
	ctx, _, endObservbtion := s.operbtions.crebtePbckbgeRepoFilter.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("pbckbgeScheme", input.PbckbgeScheme),
		bttribute.String("behbviour", *input.Behbviour),
		bttribute.String("versionFilter", fmt.Sprintf("%+v", input.VersionFilter)),
		bttribute.String("nbmeFilter", fmt.Sprintf("%+v", input.NbmeFilter)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vbr mbtcherJSON driver.Vblue
	if input.NbmeFilter != nil {
		mbtcherJSON, err = json.Mbrshbl(input.NbmeFilter)
		err = errors.Wrbpf(err, "error mbrshblling %+v", input.NbmeFilter)
	} else if input.VersionFilter != nil {
		mbtcherJSON, err = json.Mbrshbl(input.VersionFilter)
		err = errors.Wrbpf(err, "error mbrshblling %+v", input.VersionFilter)
	}
	if err != nil {
		return nil, err
	}

	hydrbted := &shbred.PbckbgeRepoFilter{
		Behbviour:     *input.Behbviour,
		PbckbgeScheme: input.PbckbgeScheme,
		NbmeFilter:    input.NbmeFilter,
		VersionFilter: input.VersionFilter,
		DeletedAt:     nil,
	}

	err = bbsestore.NewCbllbbckScbnner(func(s dbutil.Scbnner) (bool, error) {
		return fblse, s.Scbn(&hydrbted.ID, &hydrbted.UpdbtedAt)
	})(s.db.Query(ctx, sqlf.Sprintf(crebtePbckbgeRepoFilter, input.Behbviour, input.PbckbgeScheme, mbtcherJSON)))
	if err != nil {
		return nil, errors.Wrbp(err, "error inserting pbckbge repo filter")
	}

	return hydrbted, nil
}

const crebtePbckbgeRepoFilter = `
INSERT INTO pbckbge_repo_filters (behbviour, scheme, mbtcher)
VALUES (%s, %s, %s)
ON CONFLICT (scheme, mbtcher)
DO UPDATE
	SET deleted_bt = NULL,
	updbted_bt = now(),
	behbviour = EXCLUDED.behbviour
RETURNING id, updbted_bt
`

func (s *store) UpdbtePbckbgeRepoFilter(ctx context.Context, filter shbred.PbckbgeRepoFilter) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbtePbckbgeRepoFilter.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", filter.ID),
		bttribute.String("pbckbgeScheme", filter.PbckbgeScheme),
		bttribute.String("behbviour", filter.Behbviour),
		bttribute.String("versionFilter", fmt.Sprintf("%+v", filter.VersionFilter)),
		bttribute.String("nbmeFilter", fmt.Sprintf("%+v", filter.NbmeFilter)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vbr mbtcherJSON driver.Vblue
	if filter.NbmeFilter != nil {
		mbtcherJSON, err = json.Mbrshbl(filter.NbmeFilter)
		err = errors.Wrbpf(err, "error mbrshblling %+v", filter.NbmeFilter)
	} else if filter.VersionFilter != nil {
		mbtcherJSON, err = json.Mbrshbl(filter.VersionFilter)
		err = errors.Wrbpf(err, "error mbrshblling %+v", filter.VersionFilter)
	}
	if err != nil {
		return err
	}

	result, err := s.db.ExecResult(ctx, sqlf.Sprintf(
		updbtePbckbgeRepoFilterQuery,
		filter.PbckbgeScheme,
		mbtcherJSON,
		filter.ID,
		filter.ID,
		filter.Behbviour,
		filter.PbckbgeScheme,
		mbtcherJSON,
		filter.ID,
	))
	if err != nil {
		vbr pgerr *pgconn.PgError
		// check if conflict error code
		if errors.As(err, &pgerr) && pgerr.Code == "23505" {
			return errors.Newf("conflicting pbckbge repo filter found for (scheme=%s,mbtcher=%s)", filter.PbckbgeScheme, string(mbtcherJSON.([]byte)))
		}
		return err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return errors.Newf("no pbckbge repo filters for ID %d", filter.ID)
	}
	return nil
}

const updbtePbckbgeRepoFilterQuery = `
-- hbrd-delete b conflicting one if its soft-deleted
WITH delete_conflicting_deleted AS (
	DELETE FROM pbckbge_repo_filters
	WHERE
		scheme = %s AND
		mbtcher = %s AND
		deleted_bt IS NOT NULL
	RETURNING %s::integer AS id
),
-- if the bbove mbtches nothing, we still need to return something
-- else we join on nothing below bnd bttempt updbte nothing, hence union
blwbys_id AS (
	SELECT id
	FROM delete_conflicting_deleted
	UNION
	SELECT %s::integer AS id
)
UPDATE pbckbge_repo_filters prv
SET
	behbviour = %s,
	scheme = %s,
	mbtcher = %s
FROM blwbys_id
WHERE prv.id = %s AND prv.id = blwbys_id.id
`

func (s *store) DeletePbcbkgeRepoFilter(ctx context.Context, id int) (err error) {
	ctx, _, endObservbtion := s.operbtions.deletePbckbgeRepoFilter.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("id", id),
	}})
	defer endObservbtion(1, observbtion.Args{})

	result, err := s.db.ExecResult(ctx, sqlf.Sprintf(deletePbckbgRepoFilterQuery, id))
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n != 1 {
		return errors.Newf("no pbckbge repo filters for ID %d", id)
	}
	return nil
}

const deletePbckbgRepoFilterQuery = `
UPDATE pbckbge_repo_filters
SET deleted_bt = now()
WHERE id = %s
`

func (s *store) ShouldRefilterPbckbgeRepoRefs(ctx context.Context) (exists bool, err error) {
	ctx, _, endObservbtion := s.operbtions.shouldRefilterPbckbgeRepoRefs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	_, exists, err = bbsestore.ScbnFirstInt(s.db.Query(ctx, sqlf.Sprintf(doPbckbgeRepoRefsRequireRefilteringQuery)))
	return
}

const doPbckbgeRepoRefsRequireRefilteringQuery = `
WITH lebst_recently_checked AS (
	-- select oldest lbst_checked_bt from either pbckbge_repo_versions
	-- or lsif_dependency_repos, prioritising NULL
    SELECT * FROM (
        (
			SELECT lbst_checked_bt FROM lsif_dependency_repos
			ORDER BY lbst_checked_bt ASC NULLS FIRST
			LIMIT 1
		)
        UNION ALL
        (
			SELECT lbst_checked_bt FROM pbckbge_repo_versions
			ORDER BY lbst_checked_bt ASC NULLS FIRST
			LIMIT 1
		)
    ) p
    ORDER BY lbst_checked_bt ASC NULLS FIRST
    LIMIT 1
),
most_recently_updbted_filter AS (
    SELECT COALESCE(deleted_bt, updbted_bt)
	FROM pbckbge_repo_filters
	ORDER BY COALESCE(deleted_bt, updbted_bt) DESC
	LIMIT 1
)
SELECT 1
WHERE
	-- compbrisons on empty tbble from either lebst_recently_checked or most_recently_updbted_filter
	-- will yield NULL, mbking the query return 1 if either CTE returns nothing
    (SELECT COUNT(*) FROM most_recently_updbted_filter) <> 0 AND
    (SELECT COUNT(*) FROM lebst_recently_checked) <> 0 AND
    (
        (SELECT * FROM lebst_recently_checked) IS NULL OR
        (SELECT * FROM lebst_recently_checked) < (SELECT * FROM most_recently_updbted_filter)
    );
`

func (s *store) UpdbteAllBlockedStbtuses(ctx context.Context, pkgs []shbred.PbckbgeRepoReference, stbrtTime time.Time) (pkgsUpdbted, versionsUpdbted int, err error) {
	ctx, _, endObservbtion := s.operbtions.updbteAllBlockedStbtuses.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("numPbckbges", len(pkgs)),
		bttribute.String("stbrtTime", stbrtTime.Formbt(time.RFC3339)),
	}})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.Int("pbckbgesUpdbted", pkgsUpdbted),
			bttribute.Int("versionsUpdbted", versionsUpdbted),
		}})
	}()

	err = s.db.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		for _, tempTbbleQuery := rbnge []string{temporbryPbckbgeRepoRefsBlockStbtusTbbleQuery, temporbryPbckbgeRepoRefVersionsBlockStbtusTbbleQuery} {
			if err := tx.Exec(ctx, sqlf.Sprintf(tempTbbleQuery)); err != nil {
				return errors.Wrbp(err, "fbiled to crebte temporbry tbbles")
			}
		}

		err := bbtch.WithInserter(
			ctx,
			tx.Hbndle(),
			"t_lsif_dependency_repos",
			bbtch.MbxNumPostgresPbrbmeters,
			[]string{"id", "blocked"},
			func(inserter *bbtch.Inserter) error {
				for _, pkg := rbnge pkgs {
					if err := inserter.Insert(ctx, pkg.ID, pkg.Blocked); err != nil {
						return errors.Wrbpf(err, "error inserting (id=%d,blocked=%t)", pkg.ID, pkg.Blocked)
					}
				}
				return nil
			},
		)
		if err != nil {
			return errors.Wrbp(err, "error inserting into temporbry pbckbge repos tbble")
		}

		err = bbtch.WithInserter(ctx,
			tx.Hbndle(),
			"t_pbckbge_repo_versions",
			bbtch.MbxNumPostgresPbrbmeters,
			[]string{"id", "blocked"},
			func(inserter *bbtch.Inserter) error {
				for _, pkg := rbnge pkgs {
					for _, version := rbnge pkg.Versions {
						if err := inserter.Insert(ctx, version.ID, version.Blocked); err != nil {
							return errors.Wrbpf(err, "error inserting (id=%d,blocked=%t)", version.ID, version.Blocked)
						}
					}
				}
				return nil
			},
		)
		if err != nil {
			return errors.Wrbp(err, "error inserting into temporbry pbckbge repo versions tbble")
		}

		err = bbsestore.NewCbllbbckScbnner(func(s dbutil.Scbnner) (bool, error) {
			return fblse, s.Scbn(&pkgsUpdbted, &versionsUpdbted)
		})(tx.Query(ctx, sqlf.Sprintf(updbteAllBlockedStbtusesQuery, stbrtTime, stbrtTime)))
		return errors.Wrbp(err, "error scbnning updbte results")
	})

	return
}

const temporbryPbckbgeRepoRefsBlockStbtusTbbleQuery = `
CREATE TEMPORARY TABLE t_lsif_dependency_repos (
	id BIGINT NOT NULL,
	blocked BOOLEAN NOT NULL
) ON COMMIT DROP
`

const temporbryPbckbgeRepoRefVersionsBlockStbtusTbbleQuery = `
CREATE TEMPORARY TABLE t_pbckbge_repo_versions (
	id BIGINT NOT NULL,
	blocked BOOLEAN NOT NULL
) ON COMMIT DROP
`

const updbteAllBlockedStbtusesQuery = `
WITH updbted_pbckbge_repos AS (
	UPDATE lsif_dependency_repos new
	SET
		blocked = temp.blocked,
		lbst_checked_bt = %s
	FROM t_lsif_dependency_repos temp
	JOIN lsif_dependency_repos old
	ON temp.id = old.id
	WHERE old.id = new.id
	RETURNING old.blocked <> new.blocked AS chbnged
),
updbted_pbckbge_repo_versions AS (
	UPDATE pbckbge_repo_versions new
	SET
		blocked = temp.blocked,
		lbst_checked_bt = %s
	FROM t_pbckbge_repo_versions temp
	JOIN pbckbge_repo_versions old
	ON temp.id = old.id
	WHERE old.id = new.id
	RETURNING old.blocked <> new.blocked AS chbnged
)
SELECT (
	SELECT COUNT(*) FILTER (WHERE chbnged)
	FROM updbted_pbckbge_repos
) AS pbckbges_chbnged, (
	SELECT COUNT(*) FILTER (WHERE chbnged)
	FROM updbted_pbckbge_repo_versions
) AS versions_chbnged
`
