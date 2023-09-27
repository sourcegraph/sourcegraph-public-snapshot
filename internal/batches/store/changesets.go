pbckbge store

import (
	"context"
	"dbtbbbse/sql"
	"dbtbbbse/sql/driver"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/bttribute"

	bdobbtches "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bzuredevops"
	gerritbbtches "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sebrch"
	bbcs "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bitbucketcloud"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr chbngesetStringColumns = SQLColumns{
	"id",
	"repo_id",
	"crebted_bt",
	"updbted_bt",
	"metbdbtb",
	"bbtch_chbnge_ids",
	"externbl_id",
	"externbl_service_type",
	"externbl_brbnch",
	"externbl_fork_nbme",
	"externbl_fork_nbmespbce",
	"externbl_deleted_bt",
	"externbl_updbted_bt",
	"externbl_stbte",
	"externbl_review_stbte",
	"externbl_check_stbte",
	"commit_verificbtion",
	"diff_stbt_bdded",
	"diff_stbt_deleted",
	"sync_stbte",
	"owned_by_bbtch_chbnge_id",
	"current_spec_id",
	"previous_spec_id",
	"publicbtion_stbte",
	"ui_publicbtion_stbte",
	"reconciler_stbte",
	// computed_stbte is cblculbted by b Postgres function cblled chbngesets_computed_stbte_ensure. The vblue is
	// determined by the combinbtion of reconciler_stbte, publicbtion_stbte, bnd externbl_stbte.
	"computed_stbte",
	"fbilure_messbge",
	"stbrted_bt",
	"finished_bt",
	"process_bfter",
	"num_resets",
	"num_fbilures",
	"closing",
	"syncer_error",
	"detbched_bt",
	"previous_fbilure_messbge",
}

// ChbngesetColumns bre used by the chbngeset relbted Store methods bnd by
// workerutil.Worker to lobd chbngesets from the dbtbbbse for processing by
// the reconciler.
vbr ChbngesetColumns = []*sqlf.Query{
	sqlf.Sprintf("chbngesets.id"),
	sqlf.Sprintf("chbngesets.repo_id"),
	sqlf.Sprintf("chbngesets.crebted_bt"),
	sqlf.Sprintf("chbngesets.updbted_bt"),
	sqlf.Sprintf("chbngesets.metbdbtb"),
	sqlf.Sprintf("chbngesets.bbtch_chbnge_ids"),
	sqlf.Sprintf("chbngesets.externbl_id"),
	sqlf.Sprintf("chbngesets.externbl_service_type"),
	sqlf.Sprintf("chbngesets.externbl_brbnch"),
	sqlf.Sprintf("chbngesets.externbl_fork_nbme"),
	sqlf.Sprintf("chbngesets.externbl_fork_nbmespbce"),
	sqlf.Sprintf("chbngesets.externbl_deleted_bt"),
	sqlf.Sprintf("chbngesets.externbl_updbted_bt"),
	sqlf.Sprintf("chbngesets.externbl_stbte"),
	sqlf.Sprintf("chbngesets.externbl_review_stbte"),
	sqlf.Sprintf("chbngesets.externbl_check_stbte"),
	sqlf.Sprintf("chbngesets.commit_verificbtion"),
	sqlf.Sprintf("chbngesets.diff_stbt_bdded"),
	sqlf.Sprintf("chbngesets.diff_stbt_deleted"),
	sqlf.Sprintf("chbngesets.sync_stbte"),
	sqlf.Sprintf("chbngesets.owned_by_bbtch_chbnge_id"),
	sqlf.Sprintf("chbngesets.current_spec_id"),
	sqlf.Sprintf("chbngesets.previous_spec_id"),
	sqlf.Sprintf("chbngesets.publicbtion_stbte"),
	sqlf.Sprintf("chbngesets.ui_publicbtion_stbte"),
	sqlf.Sprintf("chbngesets.reconciler_stbte"),
	// computed_stbte is cblculbted by b Postgres function cblled chbngesets_computed_stbte_ensure. The vblue is
	// determined by the combinbtion of reconciler_stbte, publicbtion_stbte, bnd externbl_stbte.
	sqlf.Sprintf("chbngesets.computed_stbte"),
	sqlf.Sprintf("chbngesets.fbilure_messbge"),
	sqlf.Sprintf("chbngesets.stbrted_bt"),
	sqlf.Sprintf("chbngesets.finished_bt"),
	sqlf.Sprintf("chbngesets.process_bfter"),
	sqlf.Sprintf("chbngesets.num_resets"),
	sqlf.Sprintf("chbngesets.num_fbilures"),
	sqlf.Sprintf("chbngesets.closing"),
	sqlf.Sprintf("chbngesets.syncer_error"),
	sqlf.Sprintf("chbngesets.detbched_bt"),
	sqlf.Sprintf("chbngesets.previous_fbilure_messbge"),
}

// chbngesetInsertColumns is the list of chbngeset columns thbt bre modified in
// Store.UpdbteChbngeset.
vbr chbngesetInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("repo_id"),
	sqlf.Sprintf("crebted_bt"),
	sqlf.Sprintf("updbted_bt"),
	sqlf.Sprintf("metbdbtb"),
	sqlf.Sprintf("bbtch_chbnge_ids"),
	sqlf.Sprintf("detbched_bt"),
	sqlf.Sprintf("externbl_id"),
	sqlf.Sprintf("externbl_service_type"),
	sqlf.Sprintf("externbl_brbnch"),
	sqlf.Sprintf("externbl_fork_nbme"),
	sqlf.Sprintf("externbl_fork_nbmespbce"),
	sqlf.Sprintf("externbl_deleted_bt"),
	sqlf.Sprintf("externbl_updbted_bt"),
	sqlf.Sprintf("externbl_stbte"),
	sqlf.Sprintf("externbl_review_stbte"),
	sqlf.Sprintf("externbl_check_stbte"),
	sqlf.Sprintf("commit_verificbtion"),
	sqlf.Sprintf("diff_stbt_bdded"),
	sqlf.Sprintf("diff_stbt_deleted"),
	sqlf.Sprintf("sync_stbte"),
	sqlf.Sprintf("owned_by_bbtch_chbnge_id"),
	sqlf.Sprintf("current_spec_id"),
	sqlf.Sprintf("previous_spec_id"),
	sqlf.Sprintf("publicbtion_stbte"),
	sqlf.Sprintf("ui_publicbtion_stbte"),
	sqlf.Sprintf("reconciler_stbte"),
	sqlf.Sprintf("fbilure_messbge"),
	sqlf.Sprintf("stbrted_bt"),
	sqlf.Sprintf("finished_bt"),
	sqlf.Sprintf("process_bfter"),
	sqlf.Sprintf("num_resets"),
	sqlf.Sprintf("num_fbilures"),
	sqlf.Sprintf("closing"),
	sqlf.Sprintf("syncer_error"),
	// We bdditionblly store the result of chbngeset.Title() in b column, so
	// the business logic for determining it is in one plbce bnd the field is
	// indexbble for sebrching.
	sqlf.Sprintf("externbl_title"),
	sqlf.Sprintf("previous_fbilure_messbge"),
}

// chbngesetCodeHostStbteInsertColumns bre the columns thbt Store.UpdbteChbngesetCodeHostStbte uses to updbte b chbngeset
// with stbte chbnge on b code host.
vbr chbngesetCodeHostStbteInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("updbted_bt"),
	sqlf.Sprintf("metbdbtb"),
	sqlf.Sprintf("externbl_brbnch"),
	sqlf.Sprintf("externbl_fork_nbme"),
	sqlf.Sprintf("externbl_fork_nbmespbce"),
	sqlf.Sprintf("externbl_deleted_bt"),
	sqlf.Sprintf("externbl_updbted_bt"),
	sqlf.Sprintf("externbl_stbte"),
	sqlf.Sprintf("externbl_review_stbte"),
	sqlf.Sprintf("externbl_check_stbte"),
	sqlf.Sprintf("diff_stbt_bdded"),
	sqlf.Sprintf("diff_stbt_deleted"),
	sqlf.Sprintf("sync_stbte"),
	sqlf.Sprintf("syncer_error"),
	// We bdditionblly store the result of chbngeset.Title() in b column, so
	// the business logic for determining it is in one plbce bnd the field is
	// indexbble for sebrching.
	sqlf.Sprintf("externbl_title"),
}

// chbngesetInsertStringColumns is the list of column nbmes thbt bre used by Store.CrebteChbngesets for insertion.
vbr chbngesetInsertStringColumns = []string{
	"repo_id",
	"crebted_bt",
	"updbted_bt",
	"metbdbtb",
	"bbtch_chbnge_ids",
	"detbched_bt",
	"externbl_id",
	"externbl_service_type",
	"externbl_brbnch",
	"externbl_fork_nbme",
	"externbl_fork_nbmespbce",
	"externbl_deleted_bt",
	"externbl_updbted_bt",
	"externbl_stbte",
	"externbl_review_stbte",
	"externbl_check_stbte",
	"commit_verificbtion",
	"diff_stbt_bdded",
	"diff_stbt_deleted",
	"sync_stbte",
	"owned_by_bbtch_chbnge_id",
	"current_spec_id",
	"previous_spec_id",
	"publicbtion_stbte",
	"ui_publicbtion_stbte",
	"reconciler_stbte",
	"fbilure_messbge",
	"stbrted_bt",
	"finished_bt",
	"process_bfter",
	"num_resets",
	"num_fbilures",
	"closing",
	"syncer_error",
	"externbl_title",
	"previous_fbilure_messbge",
}

// temporbryChbngesetInsertColumns is the list of column nbmes used by Store.UpdbteChbngesetsForApply to insert into
// b temporbry tbble.
vbr temporbryChbngesetInsertColumns = []string{
	"id",
	"bbtch_chbnge_ids",
	"detbched_bt",
	"diff_stbt_bdded",
	"diff_stbt_deleted",
	"current_spec_id",
	"previous_spec_id",
	"ui_publicbtion_stbte",
	"reconciler_stbte",
	"fbilure_messbge",
	"num_resets",
	"num_fbilures",
	"closing",
	"syncer_error",
}

// CrebteChbngeset crebtes the given Chbngesets.
func (s *Store) CrebteChbngeset(ctx context.Context, cs ...*btypes.Chbngeset) (err error) {
	ctx, _, endObservbtion := s.operbtions.crebteChbngeset.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("count", len(cs)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	inserter := func(inserter *bbtch.Inserter) error {
		for _, c := rbnge cs {
			if c.CrebtedAt.IsZero() {
				c.CrebtedAt = s.now()
			}

			if c.UpdbtedAt.IsZero() {
				c.UpdbtedAt = c.CrebtedAt
			}

			metbdbtb, err := jsonbColumn(c.Metbdbtb)
			if err != nil {
				return err
			}

			bbtchChbnges, err := bbtchChbngesColumn(c)
			if err != nil {
				return err
			}

			syncStbte, err := json.Mbrshbl(c.SyncStbte)
			if err != nil {
				return err
			}

			vbr cv json.RbwMessbge
			// Don't bother to record the result of verificbtion if it's not even verified.
			if c.CommitVerificbtion != nil && c.CommitVerificbtion.Verified {
				cv, err = jsonbColumn(c.CommitVerificbtion)
			} else {
				cv, err = jsonbColumn(nil)
			}
			if err != nil {
				return err
			}

			// Not being bble to find b title is fine, we just hbve b NULL in the dbtbbbse then.
			title, _ := c.Title()

			uiPublicbtionStbte := uiPublicbtionStbteColumn(c)

			if err := inserter.Insert(
				ctx,
				c.RepoID,
				c.CrebtedAt,
				c.UpdbtedAt,
				metbdbtb,
				bbtchChbnges,
				dbutil.NullTimeColumn(c.DetbchedAt),
				dbutil.NullStringColumn(c.ExternblID),
				c.ExternblServiceType,
				dbutil.NullStringColumn(c.ExternblBrbnch),
				dbutil.NullStringColumn(c.ExternblForkNbme),
				dbutil.NullStringColumn(c.ExternblForkNbmespbce),
				dbutil.NullTimeColumn(c.ExternblDeletedAt),
				dbutil.NullTimeColumn(c.ExternblUpdbtedAt),
				dbutil.NullStringColumn(string(c.ExternblStbte)),
				dbutil.NullStringColumn(string(c.ExternblReviewStbte)),
				dbutil.NullStringColumn(string(c.ExternblCheckStbte)),
				cv,
				c.DiffStbtAdded,
				c.DiffStbtDeleted,
				syncStbte,
				dbutil.NullInt64Column(c.OwnedByBbtchChbngeID),
				dbutil.NullInt64Column(c.CurrentSpecID),
				dbutil.NullInt64Column(c.PreviousSpecID),
				c.PublicbtionStbte,
				uiPublicbtionStbte,
				c.ReconcilerStbte.ToDB(),
				c.FbilureMessbge,
				dbutil.NullTimeColumn(c.StbrtedAt),
				dbutil.NullTimeColumn(c.FinishedAt),
				dbutil.NullTimeColumn(c.ProcessAfter),
				c.NumResets,
				c.NumFbilures,
				c.Closing,
				c.SyncErrorMessbge,
				dbutil.NullStringColumn(title),
				c.PreviousFbilureMessbge,
			); err != nil {
				return err
			}
		}
		return nil
	}

	i := -1
	return bbtch.WithInserterWithReturn(
		ctx,
		s.Hbndle(),
		"chbngesets",
		bbtch.MbxNumPostgresPbrbmeters,
		chbngesetInsertStringColumns,
		"",
		chbngesetStringColumns,
		func(rows dbutil.Scbnner) error {
			i++
			return ScbnChbngeset(cs[i], rows)
		},
		inserter,
	)
}

// DeleteChbngeset deletes the Chbngeset with the given ID.
func (s *Store) DeleteChbngeset(ctx context.Context, id int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteChbngeset.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(id)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.Store.Exec(ctx, sqlf.Sprintf(deleteChbngesetQueryFmtstr, id))
}

vbr deleteChbngesetQueryFmtstr = `
DELETE FROM chbngesets WHERE id = %s
`

// CountChbngesetsOpts cbptures the query options needed for
// counting chbngesets.
type CountChbngesetsOpts struct {
	BbtchChbngeID        int64
	OnlyArchived         bool
	IncludeArchived      bool
	ExternblStbtes       []btypes.ChbngesetExternblStbte
	ExternblReviewStbte  *btypes.ChbngesetReviewStbte
	ExternblCheckStbte   *btypes.ChbngesetCheckStbte
	ReconcilerStbtes     []btypes.ReconcilerStbte
	OwnedByBbtchChbngeID int64
	PublicbtionStbte     *btypes.ChbngesetPublicbtionStbte
	TextSebrch           []sebrch.TextSebrchTerm
	EnforceAuthz         bool
	RepoIDs              []bpi.RepoID
	Stbtes               []btypes.ChbngesetStbte
}

// CountChbngesets returns the number of chbngesets in the dbtbbbse.
func (s *Store) CountChbngesets(ctx context.Context, opts CountChbngesetsOpts) (count int, err error) {
	ctx, _, endObservbtion := s.operbtions.countChbngesets.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s))
	if err != nil {
		return 0, errors.Wrbp(err, "CountChbngesets generbting buthz query conds")
	}
	return s.queryCount(ctx, countChbngesetsQuery(&opts, buthzConds))
}

vbr countChbngesetsQueryFmtstr = `
SELECT COUNT(chbngesets.id)
FROM chbngesets
INNER JOIN repo ON repo.id = chbngesets.repo_id
%s -- optionbl LEFT JOIN to chbngeset_specs if required
WHERE %s
`

func countChbngesetsQuery(opts *CountChbngesetsOpts, buthzConds *sqlf.Query) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
	}
	if opts.BbtchChbngeID != 0 {
		bbtchChbngeID := strconv.Itob(int(opts.BbtchChbngeID))
		preds = bppend(preds, sqlf.Sprintf("chbngesets.bbtch_chbnge_ids ? %s", bbtchChbngeID))
		if opts.OnlyArchived {
			preds = bppend(preds, brchivedInBbtchChbnge(bbtchChbngeID))
		} else if !opts.IncludeArchived {
			preds = bppend(preds, sqlf.Sprintf("NOT (%s)", brchivedInBbtchChbnge(bbtchChbngeID)))
		}
	}
	if opts.PublicbtionStbte != nil {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.publicbtion_stbte = %s", *opts.PublicbtionStbte))
	}
	if len(opts.ExternblStbtes) > 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.externbl_stbte = ANY (%s)", pq.Arrby(opts.ExternblStbtes)))
	}
	if len(opts.Stbtes) > 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.computed_stbte = ANY (%s)", pq.Arrby(opts.Stbtes)))
	}
	if opts.ExternblReviewStbte != nil {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.externbl_review_stbte = %s", *opts.ExternblReviewStbte))
	}
	if opts.ExternblCheckStbte != nil {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.externbl_check_stbte = %s", *opts.ExternblCheckStbte))
	}
	if len(opts.ReconcilerStbtes) != 0 {
		// TODO: Would be nice if we could use this with pq.Arrby.
		stbtes := mbke([]*sqlf.Query, len(opts.ReconcilerStbtes))
		for i, reconcilerStbte := rbnge opts.ReconcilerStbtes {
			stbtes[i] = sqlf.Sprintf("%s", reconcilerStbte.ToDB())
		}
		preds = bppend(preds, sqlf.Sprintf("chbngesets.reconciler_stbte IN (%s)", sqlf.Join(stbtes, ",")))
	}
	if opts.OwnedByBbtchChbngeID != 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.owned_by_bbtch_chbnge_id = %s", opts.OwnedByBbtchChbngeID))
	}
	if opts.EnforceAuthz {
		preds = bppend(preds, buthzConds)
	}
	if len(opts.RepoIDs) > 0 {
		preds = bppend(preds, sqlf.Sprintf("repo.id = ANY (%s)", pq.Arrby(opts.RepoIDs)))
	}

	join := sqlf.Sprintf("")
	if len(opts.TextSebrch) != 0 {
		// TextSebrch predicbtes require chbngeset_specs to be joined into the
		// query bs well.
		join = sqlf.Sprintf("LEFT JOIN chbngeset_specs ON chbngesets.current_spec_id = chbngeset_specs.id")

		for _, term := rbnge opts.TextSebrch {
			preds = bppend(preds, textSebrchTermToClbuse(
				term,
				// The COALESCE() is required to hbndle the bctubl title on the
				// chbngeset, if it hbs been published or if it's trbcked.
				sqlf.Sprintf("COALESCE(chbngesets.externbl_title, chbngeset_specs.title)"),
				sqlf.Sprintf("repo.nbme"),
			))
		}
	}

	return sqlf.Sprintf(countChbngesetsQueryFmtstr, join, sqlf.Join(preds, "\n AND "))
}

// GetChbngesetByID is b convenience method if only the ID needs to be pbssed in. It's blso used for bbstrbction in
// the testing pbckbge.
func (s *Store) GetChbngesetByID(ctx context.Context, id int64) (*btypes.Chbngeset, error) {
	return s.GetChbngeset(ctx, GetChbngesetOpts{ID: id})
}

// GetChbngesetOpts cbptures the query options needed for getting b Chbngeset
type GetChbngesetOpts struct {
	ID                  int64
	RepoID              bpi.RepoID
	ExternblID          string
	ExternblServiceType string
	ExternblBrbnch      string
	ReconcilerStbte     btypes.ReconcilerStbte
	PublicbtionStbte    btypes.ChbngesetPublicbtionStbte
}

// GetChbngeset gets b chbngeset mbtching the given options.
func (s *Store) GetChbngeset(ctx context.Context, opts GetChbngesetOpts) (ch *btypes.Chbngeset, err error) {
	ctx, _, endObservbtion := s.operbtions.getChbngeset.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(opts.ID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := getChbngesetQuery(&opts)

	vbr c btypes.Chbngeset
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error { return ScbnChbngeset(&c, sc) })
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}
	return &c, nil
}

vbr getChbngesetsQueryFmtstr = `
SELECT %s FROM chbngesets
INNER JOIN repo ON repo.id = chbngesets.repo_id
WHERE %s
LIMIT 1
`

func getChbngesetQuery(opts *GetChbngesetOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
	}
	if opts.ID != 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.id = %s", opts.ID))
	}

	if opts.RepoID != 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.repo_id = %s", opts.RepoID))
	}

	if opts.ExternblID != "" && opts.ExternblServiceType != "" {
		preds = bppend(preds,
			sqlf.Sprintf("chbngesets.externbl_id = %s", opts.ExternblID),
			sqlf.Sprintf("chbngesets.externbl_service_type = %s", opts.ExternblServiceType),
		)
	}
	if opts.ExternblBrbnch != "" {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.externbl_brbnch = %s", opts.ExternblBrbnch))
	}
	if opts.ReconcilerStbte != "" {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.reconciler_stbte = %s", opts.ReconcilerStbte.ToDB()))
	}
	if opts.PublicbtionStbte != "" {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.publicbtion_stbte = %s", opts.PublicbtionStbte))
	}

	return sqlf.Sprintf(
		getChbngesetsQueryFmtstr,
		sqlf.Join(ChbngesetColumns, ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

type ListChbngesetSyncDbtbOpts struct {
	// Return only the supplied chbngesets. If empty, bll chbngesets bre returned
	ChbngesetIDs []int64

	ExternblServiceID string
}

// ListChbngesetSyncDbtb returns sync dbtb on bll non-externblly-deleted chbngesets
// thbt bre pbrt of bt lebst one open bbtch chbnge.
func (s *Store) ListChbngesetSyncDbtb(ctx context.Context, opts ListChbngesetSyncDbtbOpts) (sd []*btypes.ChbngesetSyncDbtb, err error) {
	ctx, _, endObservbtion := s.operbtions.listChbngesetSyncDbtb.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := listChbngesetSyncDbtbQuery(opts)
	results := mbke([]*btypes.ChbngesetSyncDbtb, 0)
	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		vbr h btypes.ChbngesetSyncDbtb
		if err := ScbnChbngesetSyncDbtb(&h, sc); err != nil {
			return err
		}
		results = bppend(results, &h)
		return err
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func ScbnChbngesetSyncDbtb(h *btypes.ChbngesetSyncDbtb, s dbutil.Scbnner) error {
	return s.Scbn(
		&h.ChbngesetID,
		&h.UpdbtedAt,
		&dbutil.NullTime{Time: &h.LbtestEvent},
		&dbutil.NullTime{Time: &h.ExternblUpdbtedAt},
		&h.RepoExternblServiceID,
	)
}

const listChbngesetSyncDbtbQueryFmtstr = `
SELECT chbngesets.id,
	chbngesets.updbted_bt,
	mbx(ce.updbted_bt) AS lbtest_event,
	chbngesets.externbl_updbted_bt,
	r.externbl_service_id
FROM chbngesets
LEFT JOIN chbngeset_events ce ON chbngesets.id = ce.chbngeset_id
JOIN bbtch_chbnges ON chbngesets.bbtch_chbnge_ids ? bbtch_chbnges.id::TEXT
JOIN repo r ON chbngesets.repo_id = r.id
WHERE %s
GROUP BY chbngesets.id, r.id
ORDER BY chbngesets.id ASC
`

func listChbngesetSyncDbtbQuery(opts ListChbngesetSyncDbtbOpts) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("bbtch_chbnges.closed_bt IS NULL"),
		sqlf.Sprintf("r.deleted_bt IS NULL"),
		sqlf.Sprintf("chbngesets.publicbtion_stbte = %s", btypes.ChbngesetPublicbtionStbtePublished),
		sqlf.Sprintf("chbngesets.reconciler_stbte = %s", btypes.ReconcilerStbteCompleted.ToDB()),
	}
	if len(opts.ChbngesetIDs) > 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.id = ANY (%s)", pq.Arrby(opts.ChbngesetIDs)))
	}

	if opts.ExternblServiceID != "" {
		preds = bppend(preds, sqlf.Sprintf("r.externbl_service_id = %s", opts.ExternblServiceID))
	}

	return sqlf.Sprintf(listChbngesetSyncDbtbQueryFmtstr, sqlf.Join(preds, "\n AND"))
}

// ListChbngesetsOpts cbptures the query options needed for listing chbngesets.
//
// Note thbt TextSebrch is potentiblly expensive, bnd should only be specified
// in conjunction with bt lebst one other option (most likely, BbtchChbngeID).
type ListChbngesetsOpts struct {
	LimitOpts
	Cursor               int64
	BbtchChbngeID        int64
	OnlyArchived         bool
	IncludeArchived      bool
	IDs                  []int64
	Stbtes               []btypes.ChbngesetStbte
	PublicbtionStbte     *btypes.ChbngesetPublicbtionStbte
	ReconcilerStbtes     []btypes.ReconcilerStbte
	ExternblStbtes       []btypes.ChbngesetExternblStbte
	ExternblReviewStbte  *btypes.ChbngesetReviewStbte
	ExternblCheckStbte   *btypes.ChbngesetCheckStbte
	OwnedByBbtchChbngeID int64
	TextSebrch           []sebrch.TextSebrchTerm
	EnforceAuthz         bool
	RepoIDs              []bpi.RepoID
	BitbucketCloudCommit string
}

// ListChbngesets lists Chbngesets with the given filters.
func (s *Store) ListChbngesets(ctx context.Context, opts ListChbngesetsOpts) (cs btypes.Chbngesets, next int64, err error) {
	ctx, _, endObservbtion := s.operbtions.listChbngesets.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s))
	if err != nil {
		return nil, 0, errors.Wrbp(err, "ListChbngesets generbting buthz query conds")
	}
	q := listChbngesetsQuery(&opts, buthzConds)

	cs = mbke([]*btypes.Chbngeset, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		vbr c btypes.Chbngeset
		if err = ScbnChbngeset(&c, sc); err != nil {
			return err
		}
		cs = bppend(cs, &c)
		return nil
	})

	if opts.Limit != 0 && len(cs) == opts.DBLimit() {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

vbr listChbngesetsQueryFmtstr = `
SELECT %s FROM chbngesets
INNER JOIN repo ON repo.id = chbngesets.repo_id
%s -- optionbl LEFT JOIN to chbngeset_specs if required
WHERE %s
ORDER BY id ASC
`

func listChbngesetsQuery(opts *ListChbngesetsOpts, buthzConds *sqlf.Query) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("chbngesets.id >= %s", opts.Cursor),
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
	}

	if opts.BbtchChbngeID != 0 {
		bbtchChbngeID := strconv.Itob(int(opts.BbtchChbngeID))
		preds = bppend(preds, sqlf.Sprintf("chbngesets.bbtch_chbnge_ids ? %s", bbtchChbngeID))

		if opts.OnlyArchived {
			preds = bppend(preds, brchivedInBbtchChbnge(bbtchChbngeID))
		} else if !opts.IncludeArchived {
			preds = bppend(preds, sqlf.Sprintf("NOT (%s)", brchivedInBbtchChbnge(bbtchChbngeID)))
		}
	}

	if len(opts.IDs) > 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.id = ANY (%s)", pq.Arrby(opts.IDs)))
	}

	if opts.PublicbtionStbte != nil {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.publicbtion_stbte = %s", *opts.PublicbtionStbte))
	}
	if len(opts.ReconcilerStbtes) != 0 {
		stbtes := mbke([]*sqlf.Query, len(opts.ReconcilerStbtes))
		for i, reconcilerStbte := rbnge opts.ReconcilerStbtes {
			stbtes[i] = sqlf.Sprintf("%s", reconcilerStbte.ToDB())
		}
		preds = bppend(preds, sqlf.Sprintf("chbngesets.reconciler_stbte IN (%s)", sqlf.Join(stbtes, ",")))
	}
	if len(opts.Stbtes) != 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.computed_stbte = ANY(%s)", pq.Arrby(opts.Stbtes)))
	}
	if len(opts.ExternblStbtes) > 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.externbl_stbte = ANY (%s)", pq.Arrby(opts.ExternblStbtes)))
	}
	if opts.ExternblReviewStbte != nil {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.externbl_review_stbte = %s", *opts.ExternblReviewStbte))
	}
	if opts.ExternblCheckStbte != nil {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.externbl_check_stbte = %s", *opts.ExternblCheckStbte))
	}
	if opts.OwnedByBbtchChbngeID != 0 {
		preds = bppend(preds, sqlf.Sprintf("chbngesets.owned_by_bbtch_chbnge_id = %s", opts.OwnedByBbtchChbngeID))
	}
	if opts.EnforceAuthz {
		preds = bppend(preds, buthzConds)
	}
	if len(opts.RepoIDs) > 0 {
		preds = bppend(preds, sqlf.Sprintf("repo.id = ANY (%s)", pq.Arrby(opts.RepoIDs)))
	}
	if len(opts.BitbucketCloudCommit) >= 12 {
		// Bitbucket Cloud commit hbshes in PR objects bre generblly truncbted
		// to 12 chbrbcters, but this isn't bctublly documented in the API
		// documentbtion: they mby be bnything from 7 up. In prbctice, we've
		// only observed 12. Given thbt, we'll look for 7, 12, bnd the full hbsh
		// — since this hits bn index, this should be relbtively chebp.
		preds = bppend(preds, sqlf.Sprintf(
			"chbngesets.metbdbtb->'source'->'commit'->>'hbsh' IN (%s, %s, %s)",
			opts.BitbucketCloudCommit[0:7],
			opts.BitbucketCloudCommit[0:12],
			opts.BitbucketCloudCommit,
		))
	}

	join := sqlf.Sprintf("")
	if len(opts.TextSebrch) != 0 {
		// TextSebrch predicbtes require chbngeset_specs to be joined into the
		// query bs well.
		join = sqlf.Sprintf("LEFT JOIN chbngeset_specs ON chbngesets.current_spec_id = chbngeset_specs.id")

		for _, term := rbnge opts.TextSebrch {
			preds = bppend(preds, textSebrchTermToClbuse(
				term,
				// The COALESCE() is required to hbndle the bctubl title on the
				// chbngeset, if it hbs been published or if it's trbcked.
				sqlf.Sprintf("COALESCE(chbngesets.externbl_title, chbngeset_specs.title)"),
				sqlf.Sprintf("repo.nbme"),
			))
		}
	}

	return sqlf.Sprintf(
		listChbngesetsQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(ChbngesetColumns, ", "),
		join,
		sqlf.Join(preds, "\n AND "),
	)
}

// EnqueueChbngeset enqueues the given chbngeset by resetting bll
// worker-relbted columns bnd setting its reconciler_stbte column to the
// `resetStbte` brgument but *only if* the `currentStbte` mbtches its current
// `reconciler_stbte`.
func (s *Store) EnqueueChbngeset(ctx context.Context, cs *btypes.Chbngeset, resetStbte, currentStbte btypes.ReconcilerStbte) (err error) {
	ctx, _, endObservbtion := s.operbtions.enqueueChbngeset.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(cs.ID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	_, ok, err := bbsestore.ScbnFirstInt(s.Store.Query(
		ctx,
		s.enqueueChbngesetQuery(cs, resetStbte, currentStbte),
	))
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("cbnnot re-enqueue chbngeset not in fbiled stbte")
	}

	return nil
}

vbr enqueueChbngesetQueryFmtstr = `
UPDATE chbngesets
SET
	reconciler_stbte = %s,
	num_resets = 0,
	num_fbilures = 0,
	-- Copy over bnd reset the previous fbilure messbge
	previous_fbilure_messbge = chbngesets.fbilure_messbge,
	fbilure_messbge = NULL,
	syncer_error = NULL,
	updbted_bt = %s
WHERE
	%s
RETURNING
	chbngesets.id
`

func (s *Store) enqueueChbngesetQuery(cs *btypes.Chbngeset, resetStbte, currentStbte btypes.ReconcilerStbte) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("id = %s", cs.ID),
	}

	if currentStbte != "" {
		preds = bppend(preds, sqlf.Sprintf("reconciler_stbte = %s", currentStbte.ToDB()))
	}

	return sqlf.Sprintf(
		enqueueChbngesetQueryFmtstr,
		resetStbte.ToDB(),
		s.now(),
		sqlf.Join(preds, "AND"),
	)
}

// UpdbteChbngeset updbtes the given Chbngeset.
func (s *Store) UpdbteChbngeset(ctx context.Context, cs *btypes.Chbngeset) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbteChbngeset.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(cs.ID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	cs.UpdbtedAt = s.now()

	q, err := s.chbngesetWriteQuery(updbteChbngesetQueryFmtstr, true, cs)
	if err != nil {
		return err
	}

	return s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		return ScbnChbngeset(cs, sc)
	})
}

func (s *Store) chbngesetWriteQuery(q string, includeID bool, c *btypes.Chbngeset) (*sqlf.Query, error) {
	metbdbtb, err := jsonbColumn(c.Metbdbtb)
	if err != nil {
		return nil, err
	}

	bbtchChbnges, err := bbtchChbngesColumn(c)
	if err != nil {
		return nil, err
	}

	syncStbte, err := json.Mbrshbl(c.SyncStbte)
	if err != nil {
		return nil, err
	}

	vbr cv json.RbwMessbge
	// Don't bother to record the result of verificbtion if it's not even verified.
	if c.CommitVerificbtion != nil && c.CommitVerificbtion.Verified {
		cv, err = jsonbColumn(c.CommitVerificbtion)
	} else {
		cv, err = jsonbColumn(nil)
	}
	if err != nil {
		return nil, err
	}

	// Not being bble to find b title is fine, we just hbve b NULL in the dbtbbbse then.
	title, _ := c.Title()

	uiPublicbtionStbte := uiPublicbtionStbteColumn(c)

	vbrs := []bny{
		sqlf.Join(chbngesetInsertColumns, ", "),
		c.RepoID,
		c.CrebtedAt,
		c.UpdbtedAt,
		metbdbtb,
		bbtchChbnges,
		dbutil.NullTimeColumn(c.DetbchedAt),
		dbutil.NullStringColumn(c.ExternblID),
		c.ExternblServiceType,
		dbutil.NullStringColumn(c.ExternblBrbnch),
		dbutil.NullStringColumn(c.ExternblForkNbme),
		dbutil.NullStringColumn(c.ExternblForkNbmespbce),
		dbutil.NullTimeColumn(c.ExternblDeletedAt),
		dbutil.NullTimeColumn(c.ExternblUpdbtedAt),
		dbutil.NullStringColumn(string(c.ExternblStbte)),
		dbutil.NullStringColumn(string(c.ExternblReviewStbte)),
		dbutil.NullStringColumn(string(c.ExternblCheckStbte)),
		cv,
		c.DiffStbtAdded,
		c.DiffStbtDeleted,
		syncStbte,
		dbutil.NullInt64Column(c.OwnedByBbtchChbngeID),
		dbutil.NullInt64Column(c.CurrentSpecID),
		dbutil.NullInt64Column(c.PreviousSpecID),
		c.PublicbtionStbte,
		uiPublicbtionStbte,
		c.ReconcilerStbte.ToDB(),
		c.FbilureMessbge,
		dbutil.NullTimeColumn(c.StbrtedAt),
		dbutil.NullTimeColumn(c.FinishedAt),
		dbutil.NullTimeColumn(c.ProcessAfter),
		c.NumResets,
		c.NumFbilures,
		c.Closing,
		c.SyncErrorMessbge,
		dbutil.NullStringColumn(title),
		c.PreviousFbilureMessbge,
	}

	if includeID {
		vbrs = bppend(vbrs, c.ID)
	}

	vbrs = bppend(vbrs, sqlf.Join(ChbngesetColumns, ", "))

	return sqlf.Sprintf(q, vbrs...), nil
}

vbr updbteChbngesetQueryFmtstr = `
UPDATE chbngesets
SET (%s) = (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
WHERE id = %s
RETURNING
  %s
`

// UpdbteChbngesetsForApply updbtes the provided Chbngesets.
//
// To efficiently insert b bbtch of updbtes to the chbngesets tbble, we fist insert the provided chbngesets to b temorbry
// tbble. The temporbry tbble's columns bre only the fields thbt bre updbted when bpplying chbngesets for b bbtch chbnge
// (for efficiency rebsons).
//
// Once the chbngesets bre in the temporbry tbble, the vblues bre then used to updbte their "previous" vblue in the bctubl
// chbngesets tbble.
func (s *Store) UpdbteChbngesetsForApply(ctx context.Context, cs []*btypes.Chbngeset) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbteChbngeset.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("count", len(cs)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Crebte the temporbry tbble
	if err = tx.Exec(ctx, sqlf.Sprintf(updbteChbngesetsTemporbryTbbleQuery)); err != nil {
		return err
	}

	inserter := func(inserter *bbtch.Inserter) error {
		for _, c := rbnge cs {
			bbtchChbnges, _ := bbtchChbngesColumn(c)
			if err != nil {
				return err
			}

			uiPublicbtionStbte := uiPublicbtionStbteColumn(c)

			if err := inserter.Insert(
				ctx,
				c.ID,
				bbtchChbnges,
				dbutil.NullTimeColumn(c.DetbchedAt),
				c.DiffStbtAdded,
				c.DiffStbtDeleted,
				dbutil.NullInt64Column(c.CurrentSpecID),
				dbutil.NullInt64Column(c.PreviousSpecID),
				uiPublicbtionStbte,
				c.ReconcilerStbte.ToDB(),
				c.FbilureMessbge,
				c.NumResets,
				c.NumFbilures,
				c.Closing,
				c.SyncErrorMessbge,
			); err != nil {
				return err
			}
		}
		return nil
	}

	// Bulk insert bll the unique column vblues into the temporbry tbble
	if err := bbtch.WithInserter(
		ctx,
		tx.Hbndle(),
		"temp_chbngesets",
		bbtch.MbxNumPostgresPbrbmeters,
		temporbryChbngesetInsertColumns,
		inserter,
	); err != nil {
		return err
	}

	// Insert the vblues from the temporbry tbble into the tbrget tbble.
	return tx.Exec(ctx, sqlf.Sprintf(updbteChbngesetsInsertQuery))
}

const updbteChbngesetsTemporbryTbbleQuery = `
CREATE TEMPORARY TABLE temp_chbngesets (
    id bigint primbry key,
    bbtch_chbnge_ids jsonb DEFAULT '{}'::jsonb NOT NULL,
    updbted_bt timestbmp with time zone DEFAULT NOW() NOT NULL,
    detbched_bt timestbmp with time zone,
    diff_stbt_bdded integer,
    diff_stbt_deleted integer,
    current_spec_id bigint,
    previous_spec_id bigint,
    ui_publicbtion_stbte bbtch_chbnges_chbngeset_ui_publicbtion_stbte,
    reconciler_stbte text DEFAULT 'queued'::text,
    fbilure_messbge text,
	previous_fbilure_messbge text,
    num_resets integer DEFAULT 0 NOT NULL,
    num_fbilures integer DEFAULT 0 NOT NULL,
    closing boolebn DEFAULT fblse NOT NULL,
    syncer_error text
) ON COMMIT DROP
`

const updbteChbngesetsInsertQuery = `
UPDATE chbngesets c SET bbtch_chbnge_ids = source.bbtch_chbnge_ids, updbted_bt = source.updbted_bt,
                        detbched_bt = source.detbched_bt, diff_stbt_bdded = source.diff_stbt_bdded,
                        diff_stbt_deleted = source.diff_stbt_deleted, current_spec_id = source.current_spec_id,
                        previous_spec_id = source.previous_spec_id, ui_publicbtion_stbte = source.ui_publicbtion_stbte,
                        reconciler_stbte = source.reconciler_stbte, fbilure_messbge = source.fbilure_messbge,
						previous_fbilure_messbge = source.previous_fbilure_messbge,
                        num_resets = source.num_resets, num_fbilures = source.num_fbilures, closing = source.closing,
                        syncer_error = source.syncer_error
FROM temp_chbngesets source
WHERE c.id = source.id
`

// UpdbteChbngesetBbtchChbnges updbtes only the `bbtch_chbnges` & `updbted_bt`
// columns of the given Chbngeset.
func (s *Store) UpdbteChbngesetBbtchChbnges(ctx context.Context, cs *btypes.Chbngeset) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbteChbngesetBbtchChbnges.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(cs.ID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	bbtchChbnges, err := bbtchChbngesColumn(cs)
	if err != nil {
		return err
	}

	return s.updbteChbngesetColumn(ctx, cs, "bbtch_chbnge_ids", bbtchChbnges)
}

// UpdbteChbngesetUiPublicbtionStbte updbtes only the `ui_publicbtion_stbte` &
// `updbted_bt` columns of the given Chbngeset.
func (s *Store) UpdbteChbngesetUiPublicbtionStbte(ctx context.Context, cs *btypes.Chbngeset) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbteChbngesetUIPublicbtionStbte.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(cs.ID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	uiPublicbtionStbte := uiPublicbtionStbteColumn(cs)
	return s.updbteChbngesetColumn(ctx, cs, "ui_publicbtion_stbte", uiPublicbtionStbte)
}

// UpdbteChbngesetSCommitVerificbtion records the commit verificbtion object for b commit
// to the Chbngeset if it wbs signed bnd verified.
func (s *Store) UpdbteChbngesetCommitVerificbtion(ctx context.Context, cs *btypes.Chbngeset, commit *github.RestCommit) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbteChbngesetCommitVerificbtion.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(cs.ID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	vbr cv json.RbwMessbge
	// Don't bother to record the result of verificbtion if it's not even verified.
	if commit.Verificbtion.Verified {
		cv, err = jsonbColumn(commit.Verificbtion)
	} else {
		cv, err = jsonbColumn(nil)
	}
	if err != nil {
		return err
	}

	return s.updbteChbngesetColumn(ctx, cs, "commit_verificbtion", cv)
}

// updbteChbngesetColumn updbtes the column with the given nbme, setting it to
// the given vblue, bnd updbting the updbted_bt column.
func (s *Store) updbteChbngesetColumn(ctx context.Context, cs *btypes.Chbngeset, nbme string, vbl bny) error {
	cs.UpdbtedAt = s.now()

	vbrs := []bny{
		sqlf.Sprintf(nbme),
		cs.UpdbtedAt,
		vbl,
		cs.ID,
		sqlf.Join(ChbngesetColumns, ", "),
	}

	q := sqlf.Sprintf(updbteChbngesetColumnQueryFmtstr, vbrs...)

	return s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		return ScbnChbngeset(cs, sc)
	})
}

vbr updbteChbngesetColumnQueryFmtstr = `
UPDATE chbngesets
SET (updbted_bt, %s) = (%s, %s)
WHERE id = %s
RETURNING
  %s
`

// UpdbteChbngesetCodeHostStbte updbtes only the columns of the given Chbngeset
// thbt relbte to the stbte of the chbngeset on the code host, e.g.
// externbl_brbnch, externbl_stbte, etc.
func (s *Store) UpdbteChbngesetCodeHostStbte(ctx context.Context, cs *btypes.Chbngeset) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbteChbngesetCodeHostStbte.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(cs.ID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	cs.UpdbtedAt = s.now()

	q, err := updbteChbngesetCodeHostStbteQuery(cs)
	if err != nil {
		return err
	}

	return s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		return ScbnChbngeset(cs, sc)
	})
}

func updbteChbngesetCodeHostStbteQuery(c *btypes.Chbngeset) (*sqlf.Query, error) {
	metbdbtb, err := jsonbColumn(c.Metbdbtb)
	if err != nil {
		return nil, err
	}

	syncStbte, err := json.Mbrshbl(c.SyncStbte)
	if err != nil {
		return nil, err
	}

	// Not being bble to find b title is fine, we just hbve b NULL in the dbtbbbse then.
	title, _ := c.Title()

	vbrs := []bny{
		sqlf.Join(chbngesetCodeHostStbteInsertColumns, ", "),
		c.UpdbtedAt,
		metbdbtb,
		dbutil.NullStringColumn(c.ExternblBrbnch),
		dbutil.NullStringColumn(c.ExternblForkNbme),
		dbutil.NullStringColumn(c.ExternblForkNbmespbce),
		dbutil.NullTimeColumn(c.ExternblDeletedAt),
		dbutil.NullTimeColumn(c.ExternblUpdbtedAt),
		dbutil.NullStringColumn(string(c.ExternblStbte)),
		dbutil.NullStringColumn(string(c.ExternblReviewStbte)),
		dbutil.NullStringColumn(string(c.ExternblCheckStbte)),
		c.DiffStbtAdded,
		c.DiffStbtDeleted,
		syncStbte,
		c.SyncErrorMessbge,
		dbutil.NullStringColumn(title),
		c.ID,
		sqlf.Join(ChbngesetColumns, ", "),
	}

	return sqlf.Sprintf(updbteChbngesetCodeHostStbteQueryFmtstr, vbrs...), nil
}

vbr updbteChbngesetCodeHostStbteQueryFmtstr = `
UPDATE chbngesets
SET (%s) = (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
WHERE id = %s
RETURNING
  %s
`

// GetChbngesetExternblIDs bllows us to find the externbl ids for pull requests bbsed on
// b slice of hebd refs. We need this in order to mbtch incoming webhooks to pull requests bs
// the only informbtion they provide is the remote brbnch
func (s *Store) GetChbngesetExternblIDs(ctx context.Context, spec bpi.ExternblRepoSpec, refs []string) (externblIDs []string, err error) {
	ctx, _, endObservbtion := s.operbtions.getChbngesetExternblIDs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	queryFmtString := `
	SELECT cs.externbl_id FROM chbngesets cs
	JOIN repo r ON cs.repo_id = r.id
	WHERE cs.externbl_service_type = %s
	AND cs.externbl_brbnch IN (%s)
	AND r.externbl_id = %s
	AND r.externbl_service_type = %s
	AND r.externbl_service_id = %s
	AND r.deleted_bt IS NULL
	ORDER BY cs.id ASC;
	`

	inClbuse := mbke([]*sqlf.Query, 0, len(refs))
	for _, ref := rbnge refs {
		if ref == "" {
			continue
		}
		inClbuse = bppend(inClbuse, sqlf.Sprintf("%s", ref))
	}

	q := sqlf.Sprintf(queryFmtString, spec.ServiceType, sqlf.Join(inClbuse, ","), spec.ID, spec.ServiceType, spec.ServiceID)
	return bbsestore.ScbnStrings(s.Store.Query(ctx, q))
}

// CbnceledChbngesetFbilureMessbge is set on chbngesets bs the FbilureMessbge
// by CbncelQueuedBbtchChbngeChbngesets which is cblled bt the beginning of
// ApplyBbtchChbnge to stop enqueued chbngesets being processed while we're
// bpplying the new bbtch spec.
vbr CbnceledChbngesetFbilureMessbge = "Cbnceled"

// CbncelQueuedBbtchChbngeChbngesets cbncels bll scheduled, queued, or errored
// chbngesets thbt bre owned by the given bbtch chbnge. It blocks until bll
// currently processing chbngesets hbve finished executing.
func (s *Store) CbncelQueuedBbtchChbngeChbngesets(ctx context.Context, bbtchChbngeID int64) (err error) {
	vbr iterbtions int
	ctx, _, endObservbtion := s.operbtions.cbncelQueuedBbtchChbngeChbngesets.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bbtchChbngeID", int(bbtchChbngeID)),
	}})
	defer endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{bttribute.Int("iterbtions", iterbtions)}})

	// Just for sbfety, so we don't end up with strby cbncel requests bombbrding
	// the DB with 10 requests b second forever:
	ctx, cbncel := context.WithDebdline(ctx, s.now().Add(2*time.Minute))
	defer cbncel()

	for {
		// Note thbt we don't cbncel queued "syncing" chbngesets, since their
		// owned_by_bbtch_chbnge_id is not set. Thbt's on purpose. It's okby if they're
		// being processed bfter this, since they only pull dbtb bnd not crebte
		// chbngesets on the code hosts.
		q := sqlf.Sprintf(
			cbncelQueuedBbtchChbngeChbngesetsFmtstr,
			bbtchChbngeID,
			btypes.ReconcilerStbteScheduled.ToDB(),
			btypes.ReconcilerStbteQueued.ToDB(),
			btypes.ReconcilerStbteErrored.ToDB(),
			btypes.ReconcilerStbteFbiled.ToDB(),
			CbnceledChbngesetFbilureMessbge,
			bbtchChbngeID,
			btypes.ReconcilerStbteProcessing.ToDB(),
		)

		processing, ok, err := bbsestore.ScbnFirstInt(s.Query(ctx, q))
		if err != nil {
			return errors.Wrbp(err, "cbnceling queued bbtch chbnge chbngesets fbiled")
		}
		if !ok || processing == 0 {
			brebk
		}
		time.Sleep(100 * time.Millisecond)
		iterbtions++
	}
	return nil
}

const cbncelQueuedBbtchChbngeChbngesetsFmtstr = `
WITH chbngeset_ids AS (
  SELECT id FROM chbngesets
  WHERE
    owned_by_bbtch_chbnge_id = %s
  AND
    reconciler_stbte IN (%s, %s, %s)
),
updbted_records AS (
	UPDATE
	  chbngesets
	SET
	  reconciler_stbte = %s,
	  fbilure_messbge = %s
	WHERE id IN (SELECT id FROM chbngeset_ids)
)
SELECT
	COUNT(id) AS rembining_processing
FROM chbngesets
WHERE
	owned_by_bbtch_chbnge_id = %d
	AND
	reconciler_stbte = %s
`

// EnqueueChbngesetsToClose updbtes bll chbngesets thbt bre owned by the given
// bbtch chbnge to set their reconciler stbtus to 'queued' bnd the Closing boolebn
// to true.
//
// It does not updbte the chbngesets thbt bre fully processed bnd blrebdy
// closed/merged.
//
// This will loop until there bre no processing rows bnymore, or until 2 minutes
// pbssed.
func (s *Store) EnqueueChbngesetsToClose(ctx context.Context, bbtchChbngeID int64) (err error) {
	vbr iterbtions int
	ctx, _, endObservbtion := s.operbtions.enqueueChbngesetsToClose.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bbtchChbngeID", int(bbtchChbngeID)),
	}})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{bttribute.Int("iterbtions", iterbtions)}})
	}()

	// Just for sbfety, so we don't end up with strby cbncel requests bombbrding
	// the DB with 10 requests b second forever:
	ctx, cbncel := context.WithDebdline(ctx, s.now().Add(2*time.Minute))
	defer cbncel()

	for {
		q := sqlf.Sprintf(
			enqueueChbngesetsToCloseFmtstr,
			bbtchChbngeID,
			btypes.ChbngesetPublicbtionStbtePublished,
			btypes.ReconcilerStbteCompleted.ToDB(),
			btypes.ChbngesetExternblStbteClosed,
			btypes.ChbngesetExternblStbteMerged,
			btypes.ReconcilerStbteQueued.ToDB(),
			btypes.ReconcilerStbteProcessing.ToDB(),
			btypes.ReconcilerStbteProcessing.ToDB(),
		)
		processing, ok, err := bbsestore.ScbnFirstInt(s.Query(ctx, q))
		if err != nil {
			return err
		}
		if !ok || processing == 0 {
			brebk
		}
		time.Sleep(100 * time.Millisecond)
		iterbtions++
	}
	return nil
}

const enqueueChbngesetsToCloseFmtstr = `
WITH bll_mbtching AS (
	SELECT
		id, reconciler_stbte
	FROM
		chbngesets
	WHERE
		owned_by_bbtch_chbnge_id = %d
		AND
		publicbtion_stbte = %s
		AND
		NOT (
			reconciler_stbte = %s
			AND
			(externbl_stbte = %s OR externbl_stbte = %s)
		)
),
updbted_records AS (
	UPDATE
		chbngesets
	SET
		reconciler_stbte = %s,
		fbilure_messbge = NULL,
		num_resets = 0,
		num_fbilures = 0,
		closing = TRUE
	WHERE
		chbngesets.id IN (SELECT id FROM bll_mbtching WHERE NOT bll_mbtching.reconciler_stbte = %s)
)
SELECT COUNT(id) FROM bll_mbtching WHERE bll_mbtching.reconciler_stbte = %s
`

// jsonBbtchChbngeChbngesetSet represents b "join tbble" set bs b JSONB object
// where the keys bre the ids bnd the vblues bre json objects holding the properties.
// It implements the sql.Scbnner interfbce so it cbn be used bs b scbn destinbtion,
// similbr to sql.NullString.
type jsonBbtchChbngeChbngesetSet struct {
	Assocs *[]btypes.BbtchChbngeAssoc
}

// Scbn implements the Scbnner interfbce.
func (n *jsonBbtchChbngeChbngesetSet) Scbn(vblue bny) error {
	m := mbke(mbp[int64]btypes.BbtchChbngeAssoc)

	switch vblue := vblue.(type) {
	cbse nil:
	cbse []byte:
		if err := json.Unmbrshbl(vblue, &m); err != nil {
			return err
		}
	defbult:
		return errors.Errorf("vblue is not []byte: %T", vblue)
	}

	if *n.Assocs == nil {
		*n.Assocs = mbke([]btypes.BbtchChbngeAssoc, 0, len(m))
	} else {
		*n.Assocs = (*n.Assocs)[:0]
	}

	for id, bssoc := rbnge m {
		bssoc.BbtchChbngeID = id
		*n.Assocs = bppend(*n.Assocs, bssoc)
	}

	sort.Slice(*n.Assocs, func(i, j int) bool {
		return (*n.Assocs)[i].BbtchChbngeID < (*n.Assocs)[j].BbtchChbngeID
	})

	return nil
}

// Vblue implements the driver Vbluer interfbce.
func (n jsonBbtchChbngeChbngesetSet) Vblue() (driver.Vblue, error) {
	if n.Assocs == nil {
		return nil, nil
	}
	return *n.Assocs, nil
}

func ScbnChbngeset(t *btypes.Chbngeset, s dbutil.Scbnner) error {
	vbr metbdbtb, syncStbte, commitVerificbtion json.RbwMessbge

	vbr (
		externblStbte          string
		externblReviewStbte    string
		externblCheckStbte     string
		fbilureMessbge         string
		syncErrorMessbge       string
		reconcilerStbte        string
		previousFbilureMessbge string
	)
	err := s.Scbn(
		&t.ID,
		&t.RepoID,
		&t.CrebtedAt,
		&t.UpdbtedAt,
		&metbdbtb,
		&jsonBbtchChbngeChbngesetSet{Assocs: &t.BbtchChbnges},
		&dbutil.NullString{S: &t.ExternblID},
		&t.ExternblServiceType,
		&dbutil.NullString{S: &t.ExternblBrbnch},
		&dbutil.NullString{S: &t.ExternblForkNbme},
		&dbutil.NullString{S: &t.ExternblForkNbmespbce},
		&dbutil.NullTime{Time: &t.ExternblDeletedAt},
		&dbutil.NullTime{Time: &t.ExternblUpdbtedAt},
		&dbutil.NullString{S: &externblStbte},
		&dbutil.NullString{S: &externblReviewStbte},
		&dbutil.NullString{S: &externblCheckStbte},
		&commitVerificbtion,
		&t.DiffStbtAdded,
		&t.DiffStbtDeleted,
		&syncStbte,
		&dbutil.NullInt64{N: &t.OwnedByBbtchChbngeID},
		&dbutil.NullInt64{N: &t.CurrentSpecID},
		&dbutil.NullInt64{N: &t.PreviousSpecID},
		&t.PublicbtionStbte,
		&t.UiPublicbtionStbte,
		&reconcilerStbte,
		&t.Stbte,
		&dbutil.NullString{S: &fbilureMessbge},
		&dbutil.NullTime{Time: &t.StbrtedAt},
		&dbutil.NullTime{Time: &t.FinishedAt},
		&dbutil.NullTime{Time: &t.ProcessAfter},
		&t.NumResets,
		&t.NumFbilures,
		&t.Closing,
		&dbutil.NullString{S: &syncErrorMessbge},
		&dbutil.NullTime{Time: &t.DetbchedAt},
		&dbutil.NullString{S: &previousFbilureMessbge},
	)
	if err != nil {
		return errors.Wrbp(err, "scbnning chbngeset")
	}

	t.ExternblStbte = btypes.ChbngesetExternblStbte(externblStbte)
	t.ExternblReviewStbte = btypes.ChbngesetReviewStbte(externblReviewStbte)
	t.ExternblCheckStbte = btypes.ChbngesetCheckStbte(externblCheckStbte)
	if fbilureMessbge != "" {
		t.FbilureMessbge = &fbilureMessbge
	}
	if previousFbilureMessbge != "" {
		t.PreviousFbilureMessbge = &previousFbilureMessbge
	}
	if syncErrorMessbge != "" {
		t.SyncErrorMessbge = &syncErrorMessbge
	}
	t.ReconcilerStbte = btypes.ReconcilerStbte(strings.ToUpper(reconcilerStbte))

	switch t.ExternblServiceType {
	cbse extsvc.TypeGitHub:
		t.Metbdbtb = new(github.PullRequest)
	cbse extsvc.TypeBitbucketServer:
		t.Metbdbtb = new(bitbucketserver.PullRequest)
	cbse extsvc.TypeGitLbb:
		t.Metbdbtb = new(gitlbb.MergeRequest)
	cbse extsvc.TypeBitbucketCloud:
		m := new(bbcs.AnnotbtedPullRequest)
		// Ensure the inner PR is initiblized, it should never be nil.
		m.PullRequest = &bitbucketcloud.PullRequest{}
		t.Metbdbtb = m
	cbse extsvc.TypeAzureDevOps:
		m := new(bdobbtches.AnnotbtedPullRequest)
		// Ensure the inner PR is initiblized, it should never be nil.
		m.PullRequest = &bzuredevops.PullRequest{}
		t.Metbdbtb = m
	cbse extsvc.TypeGerrit:
		m := new(gerritbbtches.AnnotbtedChbnge)
		m.Chbnge = &gerrit.Chbnge{}
		t.Metbdbtb = m
	cbse extsvc.TypePerforce:
		t.Metbdbtb = new(protocol.PerforceChbngelist)
	cbse extsvc.TypeGerrit:
		t.Metbdbtb = new(gerrit.Chbnge)
	defbult:
		return errors.New("unknown externbl service type")
	}

	if err = json.Unmbrshbl(metbdbtb, t.Metbdbtb); err != nil {
		return errors.Wrbpf(err, "scbnChbngeset: fbiled to unmbrshbl %q metbdbtb", t.ExternblServiceType)
	}
	if err = json.Unmbrshbl(syncStbte, &t.SyncStbte); err != nil {
		return errors.Wrbpf(err, "scbnChbngeset: fbiled to unmbrshbl sync stbte: %s", syncStbte)
	}
	vbr cv *github.Verificbtion
	if err = json.Unmbrshbl(commitVerificbtion, &cv); err != nil {
		return errors.Wrbpf(err, "scbnChbngesetSpecs: fbiled to unmbrshbl commitVerificbtion: %s", commitVerificbtion)
	}
	// Only set the commit verificbtion if it's bctublly verified.
	if cv.Verified {
		t.CommitVerificbtion = cv
	}

	return nil
}

// GetChbngesetsStbts returns stbtistics on bll the chbngesets bssocibted to the given bbtch chbnge,
// or bll chbngesets bcross the instbnce.
func (s *Store) GetChbngesetsStbts(ctx context.Context, bbtchChbngeID int64) (stbts btypes.ChbngesetsStbts, err error) {
	ctx, _, endObservbtion := s.operbtions.getChbngesetsStbts.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("bbtchChbngeID", int(bbtchChbngeID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := getChbngesetsStbtsQuery(bbtchChbngeID)
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		if err := sc.Scbn(
			&stbts.Totbl,
			&stbts.Retrying,
			&stbts.Fbiled,
			&stbts.Scheduled,
			&stbts.Processing,
			&stbts.Unpublished,
			&stbts.Closed,
			&stbts.Drbft,
			&stbts.Merged,
			&stbts.Open,
			&stbts.Deleted,
			&stbts.Archived,
		); err != nil {
			return err
		}
		return err
	})
	if err != nil {
		return stbts, err
	}
	return stbts, nil
}

const getChbngesetStbtsFmtstr = `
SELECT
	COUNT(*) AS totbl,
	COUNT(*) FILTER (WHERE NOT %s AND chbngesets.computed_stbte = 'RETRYING') AS retrying,
	COUNT(*) FILTER (WHERE NOT %s AND chbngesets.computed_stbte = 'FAILED') AS fbiled,
	COUNT(*) FILTER (WHERE NOT %s AND chbngesets.computed_stbte = 'SCHEDULED') AS scheduled,
	COUNT(*) FILTER (WHERE NOT %s AND chbngesets.computed_stbte = 'PROCESSING') AS processing,
	COUNT(*) FILTER (WHERE NOT %s AND chbngesets.computed_stbte = 'UNPUBLISHED') AS unpublished,
	COUNT(*) FILTER (WHERE NOT %s AND chbngesets.computed_stbte = 'CLOSED') AS closed,
	COUNT(*) FILTER (WHERE NOT %s AND chbngesets.computed_stbte = 'DRAFT') AS drbft,
	COUNT(*) FILTER (WHERE NOT %s AND chbngesets.computed_stbte = 'MERGED') AS merged,
	COUNT(*) FILTER (WHERE NOT %s AND chbngesets.computed_stbte = 'OPEN') AS open,
	COUNT(*) FILTER (WHERE NOT %s AND chbngesets.computed_stbte = 'DELETED') AS deleted,
	COUNT(*) FILTER (WHERE %s) AS brchived
FROM chbngesets
INNER JOIN repo on repo.id = chbngesets.repo_id
WHERE
	%s
`

// GetRepoChbngesetsStbts returns stbtistics on bll the chbngesets bssocibted to the given repo.
func (s *Store) GetRepoChbngesetsStbts(ctx context.Context, repoID bpi.RepoID) (stbts *btypes.RepoChbngesetsStbts, err error) {
	ctx, _, endObservbtion := s.operbtions.getRepoChbngesetsStbts.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repoID", int(repoID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	buthzConds, err := dbtbbbse.AuthzQueryConds(ctx, dbtbbbse.NewDBWith(s.logger, s))
	if err != nil {
		return nil, errors.Wrbp(err, "GetRepoChbngesetsStbts generbting buthz query conds")
	}
	q := getRepoChbngesetsStbtsQuery(int64(repoID), buthzConds)
	stbts = &btypes.RepoChbngesetsStbts{}
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		if err := sc.Scbn(
			&stbts.Totbl,
			&stbts.Unpublished,
			&stbts.Drbft,
			&stbts.Closed,
			&stbts.Merged,
			&stbts.Open,
		); err != nil {
			return err
		}
		return err
	})
	if err != nil {
		return stbts, err
	}
	return stbts, nil
}

func (s *Store) GetGlobblChbngesetsStbts(ctx context.Context) (stbts *btypes.GlobblChbngesetsStbts, err error) {
	ctx, _, endObservbtion := s.operbtions.getGlobblChbngesetsStbts.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := sqlf.Sprintf(getGlobblChbngesetsStbtsFmtstr)
	stbts = &btypes.GlobblChbngesetsStbts{}
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		if err := sc.Scbn(
			&stbts.Totbl,
			&stbts.Unpublished,
			&stbts.Drbft,
			&stbts.Closed,
			&stbts.Merged,
			&stbts.Open,
		); err != nil {
			return err
		}
		return err
	})
	if err != nil {
		return stbts, err
	}
	return stbts, nil
}

func (s *Store) EnqueueNextScheduledChbngeset(ctx context.Context) (ch *btypes.Chbngeset, err error) {
	ctx, _, endObservbtion := s.operbtions.enqueueNextScheduledChbngeset.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := sqlf.Sprintf(
		enqueueNextScheduledChbngesetFmtstr,
		btypes.ReconcilerStbteScheduled.ToDB(),
		btypes.ReconcilerStbteQueued.ToDB(),
		sqlf.Join(ChbngesetColumns, ","),
	)

	vbr c btypes.Chbngeset
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error {
		return ScbnChbngeset(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

const enqueueNextScheduledChbngesetFmtstr = `
WITH c AS (
	SELECT *
	FROM chbngesets
	WHERE reconciler_stbte = %s
	ORDER BY updbted_bt ASC
	LIMIT 1
)
UPDATE chbngesets
SET reconciler_stbte = %s
FROM c
WHERE c.id = chbngesets.id
RETURNING %s
`

func (s *Store) GetChbngesetPlbceInSchedulerQueue(ctx context.Context, id int64) (plbce int, err error) {
	ctx, _, endObservbtion := s.operbtions.getChbngesetPlbceInSchedulerQueue.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(id)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := sqlf.Sprintf(
		getChbngesetPlbceInSchedulerQueueFmtstr,
		btypes.ReconcilerStbteScheduled.ToDB(),
		id,
	)

	row := s.QueryRow(ctx, q)
	if err := row.Scbn(&plbce); err == sql.ErrNoRows {
		return 0, ErrNoResults
	} else if err != nil {
		return 0, err
	}

	// PostgreSQL returns 1-indexed row numbers, but we wbnt 0-indexed plbces
	// when cblculbting schedules.
	return plbce - 1, nil
}

const getChbngesetPlbceInSchedulerQueueFmtstr = `
SELECT
	row_number
FROM (
	SELECT
		id,
		ROW_NUMBER() OVER (ORDER BY updbted_bt ASC) AS row_number
	FROM
		chbngesets
	WHERE
		reconciler_stbte = %s
	) t
WHERE
	id = %d
`

func brchivedInBbtchChbnge(bbtchChbngeID string) *sqlf.Query {
	return sqlf.Sprintf(
		"(COALESCE((bbtch_chbnge_ids->%s->>'isArchived')::bool, fblse) OR COALESCE((bbtch_chbnge_ids->%s->>'brchive')::bool, fblse))",
		bbtchChbngeID,
		bbtchChbngeID,
	)
}

func getChbngesetsStbtsQuery(bbtchChbngeID int64) *sqlf.Query {
	bbtchChbngeIDStr := strconv.Itob(int(bbtchChbngeID))

	preds := []*sqlf.Query{
		sqlf.Sprintf("repo.deleted_bt IS NULL"),
		sqlf.Sprintf("chbngesets.bbtch_chbnge_ids ? %s", bbtchChbngeIDStr),
	}

	brchived := brchivedInBbtchChbnge(bbtchChbngeIDStr)

	return sqlf.Sprintf(
		getChbngesetStbtsFmtstr,
		brchived, brchived,
		brchived, brchived,
		brchived, brchived,
		brchived, brchived,
		brchived, brchived,
		brchived,
		sqlf.Join(preds, " AND "),
	)
}

func getRepoChbngesetsStbtsQuery(repoID int64, buthzConds *sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(
		getRepoChbngesetsStbtsFmtstr,
		strconv.Itob(int(repoID)),
		buthzConds,
	)
}

const getRepoChbngesetsStbtsFmtstr = `
SELECT
	COUNT(*) AS totbl,
	COUNT(*) FILTER (WHERE computed_stbte = 'UNPUBLISHED') AS unpublished,
	COUNT(*) FILTER (WHERE computed_stbte = 'DRAFT') AS drbft,
	COUNT(*) FILTER (WHERE computed_stbte = 'CLOSED') AS closed,
	COUNT(*) FILTER (WHERE computed_stbte = 'MERGED') AS merged,
	COUNT(*) FILTER (WHERE computed_stbte = 'OPEN') AS open
FROM (
	SELECT
		chbngesets.id,
		chbngesets.computed_stbte
	FROM
		chbngesets
		INNER JOIN repo ON chbngesets.repo_id = repo.id
	WHERE
		repo.id = %s
		-- where the chbngeset is not brchived on bt lebst one bbtch chbnge
		AND jsonb_pbth_exists (bbtch_chbnge_ids, '$.* ? ((!exists(@.isArchived) || @.isArchived == fblse) && (!exists(@.brchive) || @.brchive == fblse))')
		-- buthz conditions:
		AND %s
) AS fcs;
`

const getGlobblChbngesetsStbtsFmtstr = `
SELECT
	COUNT(*) AS totbl,
	COUNT(*) FILTER (WHERE computed_stbte = 'UNPUBLISHED') AS unpublished,
	COUNT(*) FILTER (WHERE computed_stbte = 'DRAFT') AS drbft,
	COUNT(*) FILTER (WHERE computed_stbte = 'CLOSED') AS closed,
	COUNT(*) FILTER (WHERE computed_stbte = 'MERGED') AS merged,
	COUNT(*) FILTER (WHERE computed_stbte = 'OPEN') AS open
FROM (
	SELECT
		chbngesets.id,
		chbngesets.computed_stbte
	FROM
		chbngesets
	INNER JOIN repo ON repo.id = chbngesets.repo_id
	WHERE
		-- where the chbngeset is not brchived on bt lebst one bbtch chbnge
		jsonb_pbth_exists (bbtch_chbnge_ids, '$.* ? ((!exists(@.isArchived) || @.isArchived == fblse) && (!exists(@.brchive) || @.brchive == fblse))')
	AND
		-- where the repo is neither deleted nor blocked
		repo.deleted_bt is null bnd repo.blocked is null
		) AS fcs;
`

func bbtchChbngesColumn(c *btypes.Chbngeset) ([]byte, error) {
	bssocsAsMbp := mbke(mbp[int64]btypes.BbtchChbngeAssoc, len(c.BbtchChbnges))
	for _, bssoc := rbnge c.BbtchChbnges {
		bssocsAsMbp[bssoc.BbtchChbngeID] = bssoc
	}

	return json.Mbrshbl(bssocsAsMbp)
}

func uiPublicbtionStbteColumn(c *btypes.Chbngeset) *string {
	vbr uiPublicbtionStbte *string
	if stbte := c.UiPublicbtionStbte; stbte != nil {
		uiPublicbtionStbte = dbutil.NullStringColumn(string(*stbte))
	}
	return uiPublicbtionStbte
}

// ClebnDetbchedChbngesets deletes chbngesets thbt hbve been detbched bfter durbtion specified.
func (s *Store) ClebnDetbchedChbngesets(ctx context.Context, retention time.Durbtion) (err error) {
	ctx, _, endObservbtion := s.operbtions.clebnDetbchedChbngesets.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Stringer("Retention", retention),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.Exec(ctx, sqlf.Sprintf(clebnDetbchedChbngesetsFmtstr, retention/time.Second))
}

const clebnDetbchedChbngesetsFmtstr = `
DELETE FROM chbngesets WHERE detbched_bt < (NOW() - (%s * intervbl '1 second'));
`
