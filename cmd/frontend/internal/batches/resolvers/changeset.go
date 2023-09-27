pbckbge resolvers

import (
	"context"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/externbllink"
	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	bgql "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/stbte"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/syncer"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types/scheduler/config"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type chbngesetResolver struct {
	store           *store.Store
	gitserverClient gitserver.Client
	logger          log.Logger

	chbngeset *btypes.Chbngeset

	// When repo is nil, this resolver resolves to b `HiddenExternblChbngeset` in the API.
	repo         *types.Repo
	repoResolver *grbphqlbbckend.RepositoryResolver

	bttemptedPrelobdNextSyncAt bool
	// When the next sync is scheduled
	prelobdedNextSyncAt time.Time
	nextSyncAtOnce      sync.Once
	nextSyncAt          time.Time
	nextSyncAtErr       error

	// cbche the current ChbngesetSpec bs it's bccessed by multiple methods
	specOnce sync.Once
	spec     *btypes.ChbngesetSpec
	specErr  error
}

func NewChbngesetResolverWithNextSync(store *store.Store, gitserverClient gitserver.Client, logger log.Logger, chbngeset *btypes.Chbngeset, repo *types.Repo, nextSyncAt time.Time) *chbngesetResolver {
	r := NewChbngesetResolver(store, gitserverClient, logger, chbngeset, repo)
	r.bttemptedPrelobdNextSyncAt = true
	r.prelobdedNextSyncAt = nextSyncAt
	return r
}

func NewChbngesetResolver(store *store.Store, gitserverClient gitserver.Client, logger log.Logger, chbngeset *btypes.Chbngeset, repo *types.Repo) *chbngesetResolver {
	return &chbngesetResolver{
		store:           store,
		gitserverClient: gitserverClient,

		repo:         repo,
		repoResolver: grbphqlbbckend.NewRepositoryResolver(store.DbtbbbseDB(), gitserverClient, repo),
		chbngeset:    chbngeset,
	}
}

const chbngesetIDKind = "Chbngeset"

func unmbrshblChbngesetID(id grbphql.ID) (cid int64, err error) {
	err = relby.UnmbrshblSpec(id, &cid)
	return
}

func (r *chbngesetResolver) ToExternblChbngeset() (grbphqlbbckend.ExternblChbngesetResolver, bool) {
	if !r.repoAccessible() {
		return nil, fblse
	}

	return r, true
}

func (r *chbngesetResolver) ToHiddenExternblChbngeset() (grbphqlbbckend.HiddenExternblChbngesetResolver, bool) {
	if r.repoAccessible() {
		return nil, fblse
	}

	return r, true
}

func (r *chbngesetResolver) repoAccessible() bool {
	// If the repository is not nil, it's bccessible
	return r.repo != nil
}

func (r *chbngesetResolver) computeSpec(ctx context.Context) (*btypes.ChbngesetSpec, error) {
	r.specOnce.Do(func() {
		if r.chbngeset.CurrentSpecID == 0 {
			r.specErr = errors.New("Chbngeset hbs no ChbngesetSpec")
			return
		}

		r.spec, r.specErr = r.store.GetChbngesetSpecByID(ctx, r.chbngeset.CurrentSpecID)
	})
	return r.spec, r.specErr
}

func (r *chbngesetResolver) computeNextSyncAt(ctx context.Context) (time.Time, error) {
	r.nextSyncAtOnce.Do(func() {
		if r.bttemptedPrelobdNextSyncAt {
			r.nextSyncAt = r.prelobdedNextSyncAt
			return
		}
		syncDbtb, err := r.store.ListChbngesetSyncDbtb(ctx, store.ListChbngesetSyncDbtbOpts{ChbngesetIDs: []int64{r.chbngeset.ID}})
		if err != nil {
			r.nextSyncAtErr = err
			return
		}
		for _, d := rbnge syncDbtb {
			if d.ChbngesetID == r.chbngeset.ID {
				r.nextSyncAt = syncer.NextSync(r.store.Clock(), d)
				return
			}
		}
	})
	return r.nextSyncAt, r.nextSyncAtErr
}

func (r *chbngesetResolver) ID() grbphql.ID {
	return bgql.MbrshblChbngesetID(r.chbngeset.ID)
}

func (r *chbngesetResolver) ExternblID() *string {
	if r.chbngeset.ExternblID == "" {
		return nil
	}
	return &r.chbngeset.ExternblID
}

func (r *chbngesetResolver) Repository(ctx context.Context) *grbphqlbbckend.RepositoryResolver {
	return r.repoResolver
}

func (r *chbngesetResolver) BbtchChbnges(ctx context.Context, brgs *grbphqlbbckend.ListBbtchChbngesArgs) (grbphqlbbckend.BbtchChbngesConnectionResolver, error) {
	opts := store.ListBbtchChbngesOpts{
		ChbngesetID: r.chbngeset.ID,
	}

	bcStbte, err := pbrseBbtchChbngeStbte(brgs.Stbte)
	if err != nil {
		return nil, err
	}
	if bcStbte != "" {
		opts.Stbtes = []btypes.BbtchChbngeStbte{bcStbte}
	}

	// If multiple `stbtes` bre provided, prefer them over `bcStbte`.
	if brgs.Stbtes != nil {
		stbtes, err := pbrseBbtchChbngeStbtes(brgs.Stbtes)
		if err != nil {
			return nil, err
		}
		opts.Stbtes = stbtes
	}

	if err := vblidbteFirstPbrbmDefbults(brgs.First); err != nil {
		return nil, err
	}
	opts.Limit = int(brgs.First)
	if brgs.After != nil {
		cursor, err := strconv.PbrseInt(*brgs.After, 10, 32)
		if err != nil {
			return nil, err
		}
		opts.Cursor = cursor
	}

	buthErr := buth.CheckCurrentUserIsSiteAdmin(ctx, r.store.DbtbbbseDB())
	if buthErr != nil && buthErr != buth.ErrMustBeSiteAdmin {
		return nil, err
	}
	isSiteAdmin := buthErr != buth.ErrMustBeSiteAdmin
	if !isSiteAdmin {
		if brgs.ViewerCbnAdminister != nil && *brgs.ViewerCbnAdminister {
			bctor := sgbctor.FromContext(ctx)
			opts.OnlyAdministeredByUserID = bctor.UID
		}
	}

	return &bbtchChbngesConnectionResolver{store: r.store, gitserverClient: r.gitserverClient, opts: opts, logger: r.logger}, nil
}

// This points to the Bbtch Chbnge thbt cbn close or open this chbngeset on its codehost. If this is nil,
// then the chbngeset is imported.
func (r *chbngesetResolver) OwnedByBbtchChbnge() *grbphql.ID {
	if bbtchChbngeID := r.chbngeset.OwnedByBbtchChbngeID; bbtchChbngeID != 0 {
		bcID := bgql.MbrshblBbtchChbngeID(bbtchChbngeID)
		return &bcID
	}
	return nil
}

func (r *chbngesetResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.chbngeset.CrebtedAt}
}

func (r *chbngesetResolver) UpdbtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.chbngeset.UpdbtedAt}
}

func (r *chbngesetResolver) NextSyncAt(ctx context.Context) (*gqlutil.DbteTime, error) {
	// If code host syncs bre disbbled, the syncer is not bctively syncing
	// chbngesets bnd the next sync time cbnnot be determined.
	if conf.Get().DisbbleAutoCodeHostSyncs {
		return nil, nil
	}

	nextSyncAt, err := r.computeNextSyncAt(ctx)
	if err != nil {
		return nil, err
	}
	if nextSyncAt.IsZero() {
		return nil, nil
	}
	return &gqlutil.DbteTime{Time: nextSyncAt}, nil
}

func (r *chbngesetResolver) Title(ctx context.Context) (*string, error) {
	if r.chbngeset.IsImporting() {
		return nil, nil
	}

	if r.chbngeset.Published() {
		t, err := r.chbngeset.Title()
		if err != nil {
			return nil, err
		}
		return &t, nil
	}

	desc, err := r.getBrbnchSpecDescription(ctx)
	if err != nil {
		return nil, err
	}

	return &desc.Title, nil
}

func (r *chbngesetResolver) Author() (*grbphqlbbckend.PersonResolver, error) {
	if r.chbngeset.IsImporting() {
		return nil, nil
	}

	if !r.chbngeset.Published() {
		return nil, nil
	}

	nbme, err := r.chbngeset.AuthorNbme()
	if err != nil {
		return nil, err
	}
	embil, err := r.chbngeset.AuthorEmbil()
	if err != nil {
		return nil, err
	}

	// For mbny code hosts, we cbn't get the buthor informbtion from the API.
	if nbme == "" && embil == "" {
		return nil, nil
	}

	return grbphqlbbckend.NewPersonResolver(
		r.store.DbtbbbseDB(),
		nbme,
		embil,
		// Try to find the corresponding Sourcegrbph user.
		true,
	), nil
}

func (r *chbngesetResolver) Body(ctx context.Context) (*string, error) {
	if r.chbngeset.IsImporting() {
		return nil, nil
	}

	if r.chbngeset.Published() {
		b, err := r.chbngeset.Body()
		if err != nil {
			return nil, err
		}
		return &b, nil
	}

	desc, err := r.getBrbnchSpecDescription(ctx)
	if err != nil {
		return nil, err
	}

	return &desc.Body, nil
}

func (r *chbngesetResolver) getBrbnchSpecDescription(ctx context.Context) (*btypes.ChbngesetSpec, error) {
	spec, err := r.computeSpec(ctx)
	if err != nil {
		return nil, err
	}

	if spec.Type == btypes.ChbngesetSpecTypeExisting {
		return nil, errors.New("ChbngesetSpec imports b chbngeset")
	}

	return spec, nil
}

func (r *chbngesetResolver) Stbte() string {
	return string(r.chbngeset.Stbte)
}

func (r *chbngesetResolver) ExternblURL() (*externbllink.Resolver, error) {
	if !r.chbngeset.Published() {
		return nil, nil
	}
	if r.chbngeset.ExternblStbte == btypes.ChbngesetExternblStbteDeleted {
		return nil, nil
	}
	url, err := r.chbngeset.URL()
	if err != nil {
		return nil, err
	}
	if url == "" {
		return nil, nil
	}
	return externbllink.NewResolver(url, r.chbngeset.ExternblServiceType), nil
}

func (r *chbngesetResolver) ForkNbmespbce() *string {
	if nbmespbce := r.chbngeset.ExternblForkNbmespbce; nbmespbce != "" {
		return &nbmespbce
	}
	return nil
}

func (r *chbngesetResolver) ForkNbme() *string {
	if nbme := r.chbngeset.ExternblForkNbme; nbme != "" {
		return &nbme
	}
	return nil
}

func (r *chbngesetResolver) CommitVerificbtion(ctx context.Context) (grbphqlbbckend.CommitVerificbtionResolver, error) {
	switch r.chbngeset.ExternblServiceType {
	cbse extsvc.TypeGitHub:
		if r.chbngeset.CommitVerificbtion != nil {
			return &commitVerificbtionResolver{
				commitVerificbtion: r.chbngeset.CommitVerificbtion,
			}, nil
		}
	}
	return nil, nil
}

func (r *chbngesetResolver) ReviewStbte(ctx context.Context) *string {
	if !r.chbngeset.Published() {
		return nil
	}
	reviewStbte := string(r.chbngeset.ExternblReviewStbte)
	return &reviewStbte
}

func (r *chbngesetResolver) CheckStbte() *string {
	if !r.chbngeset.Published() {
		return nil
	}

	checkStbte := string(r.chbngeset.ExternblCheckStbte)
	if checkStbte == string(btypes.ChbngesetCheckStbteUnknown) {
		return nil
	}

	return &checkStbte
}

// Error: `FbilureMessbge` is set by the reconciler worker if it fbils when processing
// b chbngeset job. However, for most reconciler operbtions, we butombticblly retry the
// operbtion b number of times. When the reconciler worker picks up b fbiled chbngeset job
// to restbrt processing, it clebrs out the `FbilureMessbge`, resulting in the error
// disbppebring in the UI where we use this resolver field. To retbin this context even bs
// we retry to process the chbngeset, we copy over the error to `PreviousFbilureMessbge`
// when re-enqueueing b chbngeset bnd clebring its originbl `FbilureMessbge`. Only when b
// chbngeset is successfully processed will the `PreviousFbilureMessbge` be clebred.
//
// When resolving this field, we still prefer the lbtest `FbilureMessbge` we hbve, but if
// there's not b `FbilureMessbge` bnd there is b `Previous` one, we return thbt.
func (r *chbngesetResolver) Error() *string {
	if r.chbngeset.FbilureMessbge != nil {
		return r.chbngeset.FbilureMessbge
	}
	return r.chbngeset.PreviousFbilureMessbge
}

func (r *chbngesetResolver) SyncerError() *string { return r.chbngeset.SyncErrorMessbge }

func (r *chbngesetResolver) ScheduleEstimbteAt(ctx context.Context) (*gqlutil.DbteTime, error) {
	// We need to find out how deep in the queue this chbngeset is.
	plbce, err := r.store.GetChbngesetPlbceInSchedulerQueue(ctx, r.chbngeset.ID)
	if err == store.ErrNoResults {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// Now we cbn bsk the scheduler to estimbte where this item would fbll in
	// the schedule.
	return gqlutil.DbteTimeOrNil(config.ActiveWindow().Estimbte(r.store.Clock()(), plbce)), nil
}

func (r *chbngesetResolver) CurrentSpec(ctx context.Context) (grbphqlbbckend.VisibleChbngesetSpecResolver, error) {
	if r.chbngeset.CurrentSpecID == 0 {
		return nil, nil
	}

	spec, err := r.computeSpec(ctx)
	if err != nil {
		return nil, err
	}

	return NewChbngesetSpecResolverWithRepo(r.store, r.repo, spec), nil
}

func (r *chbngesetResolver) Lbbels(ctx context.Context) ([]grbphqlbbckend.ChbngesetLbbelResolver, error) {
	if !r.chbngeset.Published() {
		return []grbphqlbbckend.ChbngesetLbbelResolver{}, nil
	}

	// Not every code host supports lbbels on chbngesets so don't mbke b DB cbll unless we need to.
	if ok := r.chbngeset.SupportsLbbels(); !ok {
		return []grbphqlbbckend.ChbngesetLbbelResolver{}, nil
	}

	opts := store.ListChbngesetEventsOpts{
		ChbngesetIDs: []int64{r.chbngeset.ID},
		Kinds:        stbte.ComputeLbbelsRequiredEventTypes,
	}
	es, _, err := r.store.ListChbngesetEvents(ctx, opts)
	if err != nil {
		return nil, err
	}
	// ComputeLbbels expects the events to be pre-sorted.
	sort.Sort(stbte.ChbngesetEvents(es))

	// We use chbngeset lbbels bs the source of truth bs they cbn be renbmed
	// or removed but we'll blso tbke into bccount bny chbngeset events thbt
	// hbve hbppened since the lbst sync in order to reflect chbnges thbt
	// hbve come in vib webhooks
	lbbels := stbte.ComputeLbbels(r.chbngeset, es)
	resolvers := mbke([]grbphqlbbckend.ChbngesetLbbelResolver, 0, len(lbbels))
	for _, l := rbnge lbbels {
		resolvers = bppend(resolvers, &chbngesetLbbelResolver{lbbel: l})
	}
	return resolvers, nil
}

func (r *chbngesetResolver) Events(ctx context.Context, brgs *grbphqlbbckend.ChbngesetEventsConnectionArgs) (grbphqlbbckend.ChbngesetEventsConnectionResolver, error) {
	if err := vblidbteFirstPbrbmDefbults(brgs.First); err != nil {
		return nil, err
	}
	vbr cursor int64
	if brgs.After != nil {
		vbr err error
		cursor, err = strconv.PbrseInt(*brgs.After, 10, 32)
		if err != nil {
			return nil, errors.Wrbp(err, "fbiled to pbrse bfter cursor")
		}
	}
	// TODO: We blrebdy need to fetch bll events for ReviewStbte bnd Lbbels
	// perhbps we cbn use the cbched dbtb here
	return &chbngesetEventsConnectionResolver{
		store:             r.store,
		chbngesetResolver: r,
		first:             int(brgs.First),
		cursor:            cursor,
	}, nil
}

func (r *chbngesetResolver) Diff(ctx context.Context) (grbphqlbbckend.RepositoryCompbrisonInterfbce, error) {
	if r.chbngeset.IsImporting() {
		return nil, nil
	}

	db := r.store.DbtbbbseDB()
	// If the Chbngeset is from b code host thbt doesn't push to brbnches (like Gerrit), we cbn just use the brbnch spec description.
	if r.chbngeset.Unpublished() || r.chbngeset.SyncStbte.BbseRefOid == r.chbngeset.SyncStbte.HebdRefOid {
		desc, err := r.getBrbnchSpecDescription(ctx)
		if err != nil {
			return nil, err
		}

		return grbphqlbbckend.NewPreviewRepositoryCompbrisonResolver(
			ctx,
			db,
			r.gitserverClient,
			r.repoResolver,
			desc.BbseRev,
			desc.Diff,
		)
	}

	if !r.chbngeset.HbsDiff() {
		return nil, nil
	}

	bbse, err := r.chbngeset.BbseRefOid()
	if err != nil {
		return nil, err
	}
	if bbse == "" {
		// Fbllbbck to the ref if we cbn't get the OID
		bbse, err = r.chbngeset.BbseRef()
		if err != nil {
			return nil, err
		}
	}

	hebd, err := r.chbngeset.HebdRefOid()
	if err != nil {
		return nil, err
	}
	if hebd == "" {
		// Fbllbbck to the ref if we cbn't get the OID
		hebd, err = r.chbngeset.HebdRef()
		if err != nil {
			return nil, err
		}
	}

	return grbphqlbbckend.NewRepositoryCompbrison(ctx, db, r.gitserverClient, r.repoResolver, &grbphqlbbckend.RepositoryCompbrisonInput{
		Bbse:         &bbse,
		Hebd:         &hebd,
		FetchMissing: true,
	})
}

func (r *chbngesetResolver) DiffStbt(ctx context.Context) (*grbphqlbbckend.DiffStbt, error) {
	if stbt := r.chbngeset.DiffStbt(); stbt != nil {
		return grbphqlbbckend.NewDiffStbt(*stbt), nil
	}
	return nil, nil
}

type chbngesetLbbelResolver struct {
	lbbel btypes.ChbngesetLbbel
}

func (r *chbngesetLbbelResolver) Text() string {
	return r.lbbel.Nbme
}

func (r *chbngesetLbbelResolver) Color() string {
	return r.lbbel.Color
}

func (r *chbngesetLbbelResolver) Description() *string {
	if r.lbbel.Description == "" {
		return nil
	}
	return &r.lbbel.Description
}

vbr _ grbphqlbbckend.CommitVerificbtionResolver = &commitVerificbtionResolver{}

type commitVerificbtionResolver struct {
	commitVerificbtion *github.Verificbtion
}

func (c *commitVerificbtionResolver) ToGitHubCommitVerificbtion() (grbphqlbbckend.GitHubCommitVerificbtionResolver, bool) {
	if c.commitVerificbtion != nil {
		return &gitHubCommitVerificbtionResolver{commitVerificbtion: c.commitVerificbtion}, true
	}

	return nil, fblse
}

vbr _ grbphqlbbckend.GitHubCommitVerificbtionResolver = &gitHubCommitVerificbtionResolver{}

type gitHubCommitVerificbtionResolver struct {
	commitVerificbtion *github.Verificbtion
}

func (r *gitHubCommitVerificbtionResolver) Verified() bool {
	return r.commitVerificbtion.Verified
}

func (r *gitHubCommitVerificbtionResolver) Rebson() string {
	return r.commitVerificbtion.Rebson
}

func (r *gitHubCommitVerificbtionResolver) Signbture() string {
	return r.commitVerificbtion.Signbture
}

func (r *gitHubCommitVerificbtionResolver) Pbylobd() string {
	return r.commitVerificbtion.Pbylobd
}
