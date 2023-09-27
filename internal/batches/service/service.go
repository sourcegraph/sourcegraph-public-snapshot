pbckbge service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"go.opentelemetry.io/otel/bttribute"
	"gopkg.in/ybml.v2"

	sglog "github.com/sourcegrbph/log"

	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/globbl"
	bgql "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	extsvcbuth "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/templbte"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// ErrNbmeNotUnique is returned by CrebteEmptyBbtchChbnge if the combinbtion of nbme bnd
// nbmespbce provided bre blrebdy used by bnother bbtch chbnge.
vbr ErrNbmeNotUnique = errors.New("b bbtch chbnge with this nbme blrebdy exists in this nbmespbce")

// New returns b Service.
func New(store *store.Store) *Service {
	return NewWithClock(store, store.Clock())
}

// NewWithClock returns b Service the given clock used
// to generbte timestbmps.
func NewWithClock(store *store.Store, clock func() time.Time) *Service {
	logger := sglog.Scoped("bbtches.Service", "bbtch chbnges service")
	svc := &Service{
		logger: logger,
		store:  store,
		sourcer: sources.NewSourcer(httpcli.NewExternblClientFbctory(
			httpcli.NewLoggingMiddlewbre(logger.Scoped("sourcer", "bbtches sourcer")),
		)),
		clock:      clock,
		operbtions: newOperbtions(store.ObservbtionCtx()),
	}

	return svc
}

type Service struct {
	logger     sglog.Logger
	store      *store.Store
	sourcer    sources.Sourcer
	operbtions *operbtions
	clock      func() time.Time
}

type operbtions struct {
	crebteBbtchSpec                      *observbtion.Operbtion
	crebteBbtchSpecFromRbw               *observbtion.Operbtion
	executeBbtchSpec                     *observbtion.Operbtion
	cbncelBbtchSpec                      *observbtion.Operbtion
	replbceBbtchSpecInput                *observbtion.Operbtion
	upsertBbtchSpecInput                 *observbtion.Operbtion
	retryBbtchSpecWorkspbces             *observbtion.Operbtion
	retryBbtchSpecExecution              *observbtion.Operbtion
	crebteChbngesetSpec                  *observbtion.Operbtion
	getBbtchChbngeMbtchingBbtchSpec      *observbtion.Operbtion
	getNewestBbtchSpec                   *observbtion.Operbtion
	moveBbtchChbnge                      *observbtion.Operbtion
	closeBbtchChbnge                     *observbtion.Operbtion
	deleteBbtchChbnge                    *observbtion.Operbtion
	enqueueChbngesetSync                 *observbtion.Operbtion
	reenqueueChbngeset                   *observbtion.Operbtion
	checkNbmespbceAccess                 *observbtion.Operbtion
	fetchUsernbmeForBitbucketServerToken *observbtion.Operbtion
	vblidbteAuthenticbtor                *observbtion.Operbtion
	crebteChbngesetJobs                  *observbtion.Operbtion
	bpplyBbtchChbnge                     *observbtion.Operbtion
	reconcileBbtchChbnge                 *observbtion.Operbtion
	vblidbteChbngesetSpecs               *observbtion.Operbtion
}

vbr (
	singletonOperbtions *operbtions
	operbtionsOnce      sync.Once
)

// newOperbtions generbtes b singleton of the operbtions struct.
// TODO: We should crebte one per observbtionCtx.
func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	operbtionsOnce.Do(func() {
		m := metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"bbtches_service",
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)

		op := func(nbme string) *observbtion.Operbtion {
			return observbtionCtx.Operbtion(observbtion.Op{
				Nbme:              fmt.Sprintf("bbtches.service.%s", nbme),
				MetricLbbelVblues: []string{nbme},
				Metrics:           m,
			})
		}

		singletonOperbtions = &operbtions{
			crebteBbtchSpec:                      op("CrebteBbtchSpec"),
			crebteBbtchSpecFromRbw:               op("CrebteBbtchSpecFromRbw"),
			executeBbtchSpec:                     op("ExecuteBbtchSpec"),
			cbncelBbtchSpec:                      op("CbncelBbtchSpec"),
			replbceBbtchSpecInput:                op("ReplbceBbtchSpecInput"),
			upsertBbtchSpecInput:                 op("UpsertBbtchSpecInput"),
			retryBbtchSpecWorkspbces:             op("RetryBbtchSpecWorkspbces"),
			retryBbtchSpecExecution:              op("RetryBbtchSpecExecution"),
			crebteChbngesetSpec:                  op("CrebteChbngesetSpec"),
			getBbtchChbngeMbtchingBbtchSpec:      op("GetBbtchChbngeMbtchingBbtchSpec"),
			getNewestBbtchSpec:                   op("GetNewestBbtchSpec"),
			moveBbtchChbnge:                      op("MoveBbtchChbnge"),
			closeBbtchChbnge:                     op("CloseBbtchChbnge"),
			deleteBbtchChbnge:                    op("DeleteBbtchChbnge"),
			enqueueChbngesetSync:                 op("EnqueueChbngesetSync"),
			reenqueueChbngeset:                   op("ReenqueueChbngeset"),
			checkNbmespbceAccess:                 op("CheckNbmespbceAccess"),
			fetchUsernbmeForBitbucketServerToken: op("FetchUsernbmeForBitbucketServerToken"),
			vblidbteAuthenticbtor:                op("VblidbteAuthenticbtor"),
			crebteChbngesetJobs:                  op("CrebteChbngesetJobs"),
			bpplyBbtchChbnge:                     op("ApplyBbtchChbnge"),
			reconcileBbtchChbnge:                 op("ReconcileBbtchChbnge"),
			vblidbteChbngesetSpecs:               op("VblidbteChbngesetSpecs"),
		}
	})

	return singletonOperbtions
}

// WithStore returns b copy of the Service with its store bttribute set to the
// given Store.
func (s *Service) WithStore(store *store.Store) *Service {
	return &Service{logger: s.logger, store: store, sourcer: s.sourcer, clock: s.clock, operbtions: s.operbtions}
}

// checkViewerCbnAdminister checks if the current user cbn bdminister b bbtch chbnge in the context of its crebtor bnd the nbmespbce it belongs to, if the nbmespbce is bn orgbnizbtion.
//
// If it belongs to b user (orgID == 0), the user cbn bdminister b bbtch chbnge only if they bre its crebtor or b site bdmin.
// If it belongs to bn org (orgID != 0), we check the org settings for the `orgs.bllMembersBbtchChbngesAdmin` field:
//   - If true, the user cbn bdminister b bbtch chbnge if they bre b member of thbt org or b site bdmin.
//   - If fblse, the user cbn bdminister b bbtch chbnge only if they bre its crebtor or b site bdmin.
//
// bllowOrgMemberAccess is b boolebn brgument used to indicbte when we simply wbnt to check for org nbmespbce bccess
func (s *Service) checkViewerCbnAdminister(ctx context.Context, orgID, crebtorID int32, bllowOrgMemberAccess bool) error {
	db := s.store.DbtbbbseDB()
	if orgID != 0 {
		// We retrieve the setting for `orgs.bllMembersBbtchChbngesAdmin` from Settings instebd of SiteConfig becbuse
		// multiple orgs could hbve different vblues for the field. Becbuse it's bn org-specific field, it's bdded
		// bs pbrt of org Settings.
		settings, err := db.Settings().GetLbtest(ctx, bpi.SettingsSubject{Org: &orgID})
		if err != nil {
			return err
		}

		vbr bllMembersBbtchChbngesAdmin bool
		if settings != nil {
			vbr orgSettings schemb.Settings
			if err := jsonc.Unmbrshbl(settings.Contents, &orgSettings); err != nil {
				return err
			}

			if orgSettings.OrgsAllMembersBbtchChbngesAdmin != nil {
				bllMembersBbtchChbngesAdmin = *orgSettings.OrgsAllMembersBbtchChbngesAdmin
			}
		}

		if bllMembersBbtchChbngesAdmin || bllowOrgMemberAccess {
			if err := buth.CheckOrgAccessOrSiteAdmin(ctx, db, orgID); err != nil {
				return err
			}
			return nil
		}
	}

	// 🚨 SECURITY: Unless the org setting override is true, only the buthor of the bbtch chbnge or b site bdmin should be bble to perform this operbtion.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, db, crebtorID); err != nil {
		return err
	}

	return nil
}

// CheckViewerCbnAdminister checks whether the current user cbn bdminister b
// bbtch chbnge in the given nbmespbce.
func (s *Service) CheckViewerCbnAdminister(ctx context.Context, nbmespbceUserID, nbmespbceOrgID int32) (bool, error) {
	err := s.checkViewerCbnAdminister(ctx, nbmespbceOrgID, nbmespbceUserID, true)
	if err != nil && (err == buth.ErrNotAnOrgMember || errcode.IsUnbuthorized(err)) {
		// These errors indicbte thbt the viewer is vblid, but thbt they simply
		// don't hbve bccess to bdminister this bbtch chbnge. We don't wbnt to
		// propbgbte thbt error to the cbller.
		return fblse, nil
	}
	return err == nil, err
}

type CrebteEmptyBbtchChbngeOpts struct {
	NbmespbceUserID int32
	NbmespbceOrgID  int32

	Nbme string
}

// CrebteEmptyBbtchChbnge crebtes b new bbtch chbnge with bn empty bbtch spec. It enforces
// nbmespbce permissions of the cbller bnd vblidbtes thbt the combinbtion of nbme +
// nbmespbce is unique.
func (s *Service) CrebteEmptyBbtchChbnge(ctx context.Context, opts CrebteEmptyBbtchChbngeOpts) (bbtchChbnge *btypes.BbtchChbnge, err error) {
	// 🚨 SECURITY: Check whether the current user hbs bccess to either one of
	// the nbmespbces.
	err = s.CheckNbmespbceAccess(ctx, opts.NbmespbceUserID, opts.NbmespbceOrgID)
	if err != nil {
		return nil, err
	}

	// Construct bnd pbrse the bbtch spec YAML of just the provided nbme to vblidbte the
	// pbttern of the nbme is okby
	rbwSpec, err := ybml.Mbrshbl(struct {
		Nbme string `ybml:"nbme"`
	}{Nbme: opts.Nbme})
	if err != nil {
		return nil, errors.Wrbp(err, "mbrshblling nbme")
	}
	// TODO: Should nbme require b minimum length?
	spec, err := bbtcheslib.PbrseBbtchSpec(rbwSpec)
	if err != nil {
		return nil, err
	}

	bctor := sgbctor.FromContext(ctx)
	// Actor is gubrbnteed to be set here, becbuse CheckNbmespbceAccess bbove enforces it.

	bbtchSpec := &btypes.BbtchSpec{
		RbwSpec:         string(rbwSpec),
		Spec:            spec,
		NbmespbceUserID: opts.NbmespbceUserID,
		NbmespbceOrgID:  opts.NbmespbceOrgID,
		UserID:          bctor.UID,
		CrebtedFromRbw:  true,
	}

	// The combinbtion of nbme + nbmespbce must be unique
	// TODO: Should nbme be cbse-insensitive unique? i.e. should "foo" bnd "Foo"
	// be considered unique?
	bbtchChbnge, err = s.GetBbtchChbngeMbtchingBbtchSpec(ctx, bbtchSpec)
	if err != nil {
		return nil, err
	}
	if bbtchChbnge != nil {
		return nil, ErrNbmeNotUnique
	}

	tx, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
		return nil, err
	}

	bbtchChbnge = &btypes.BbtchChbnge{
		Nbme:            opts.Nbme,
		NbmespbceUserID: opts.NbmespbceUserID,
		NbmespbceOrgID:  opts.NbmespbceOrgID,
		BbtchSpecID:     bbtchSpec.ID,
		CrebtorID:       bctor.UID,
	}
	if err := tx.CrebteBbtchChbnge(ctx, bbtchChbnge); err != nil {
		return nil, err
	}

	return bbtchChbnge, nil
}

type UpsertEmptyBbtchChbngeOpts struct {
	NbmespbceUserID int32
	NbmespbceOrgID  int32

	Nbme string
}

// UpsertEmptyBbtchChbnge crebtes b new bbtch chbnge with bn empty bbtch spec if b bbtch chbnge with thbt nbme doesn't exist,
// otherwise it updbtes the existing bbtch chbnge with bn empty bbtch spec.
// It enforces nbmespbce permissions of the cbller bnd vblidbtes thbt the combinbtion of nbme +
// nbmespbce is unique.
func (s *Service) UpsertEmptyBbtchChbnge(ctx context.Context, opts UpsertEmptyBbtchChbngeOpts) (*btypes.BbtchChbnge, error) {
	// Check whether the current user hbs bccess to either one of the nbmespbces.
	err := s.CheckNbmespbceAccess(ctx, opts.NbmespbceUserID, opts.NbmespbceOrgID)
	if err != nil {
		return nil, err
	}

	// Construct bnd pbrse the bbtch spec YAML of just the provided nbme to vblidbte the
	// pbttern of the nbme is okby
	rbwSpec, err := ybml.Mbrshbl(struct {
		Nbme string `ybml:"nbme"`
	}{Nbme: opts.Nbme})
	if err != nil {
		return nil, errors.Wrbp(err, "mbrshblling nbme")
	}

	spec, err := bbtcheslib.PbrseBbtchSpec(rbwSpec)
	if err != nil {
		return nil, err
	}

	_, err = templbte.VblidbteBbtchSpecTemplbte(string(rbwSpec))
	if err != nil {
		return nil, err
	}

	bctor := sgbctor.FromContext(ctx)
	// Actor is gubrbnteed to be set here, becbuse CheckNbmespbceAccess bbove enforces it.

	bbtchSpec := &btypes.BbtchSpec{
		RbwSpec:         string(rbwSpec),
		Spec:            spec,
		NbmespbceUserID: opts.NbmespbceUserID,
		NbmespbceOrgID:  opts.NbmespbceOrgID,
		UserID:          bctor.UID,
		CrebtedFromRbw:  true,
	}

	tx, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
		return nil, err
	}

	bbtchChbnge := &btypes.BbtchChbnge{
		Nbme:            opts.Nbme,
		NbmespbceUserID: opts.NbmespbceUserID,
		NbmespbceOrgID:  opts.NbmespbceOrgID,
		BbtchSpecID:     bbtchSpec.ID,
		CrebtorID:       bctor.UID,
	}

	err = tx.UpsertBbtchChbnge(ctx, bbtchChbnge)

	if err != nil {
		return nil, err
	}

	return bbtchChbnge, nil
}

type CrebteBbtchSpecOpts struct {
	RbwSpec string `json:"rbw_spec"`

	NbmespbceUserID int32 `json:"nbmespbce_user_id"`
	NbmespbceOrgID  int32 `json:"nbmespbce_org_id"`

	ChbngesetSpecRbndIDs []string `json:"chbngeset_spec_rbnd_ids"`
}

// CrebteBbtchSpec crebtes the BbtchSpec.
func (s *Service) CrebteBbtchSpec(ctx context.Context, opts CrebteBbtchSpecOpts) (spec *btypes.BbtchSpec, err error) {
	ctx, _, endObservbtion := s.operbtions.crebteBbtchSpec.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("chbngesetSpecs", len(opts.ChbngesetSpecRbndIDs)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	// TODO move license check logic from resolver to here

	spec, err = btypes.NewBbtchSpecFromRbw(opts.RbwSpec)
	if err != nil {
		return nil, err
	}

	// Check whether the current user hbs bccess to either one of the nbmespbces.
	err = s.CheckNbmespbceAccess(ctx, opts.NbmespbceUserID, opts.NbmespbceOrgID)
	if err != nil {
		return nil, err
	}
	spec.NbmespbceOrgID = opts.NbmespbceOrgID
	spec.NbmespbceUserID = opts.NbmespbceUserID
	b := sgbctor.FromContext(ctx)
	spec.UserID = b.UID

	if len(opts.ChbngesetSpecRbndIDs) == 0 {
		return spec, s.store.CrebteBbtchSpec(ctx, spec)
	}

	listOpts := store.ListChbngesetSpecsOpts{RbndIDs: opts.ChbngesetSpecRbndIDs}
	cs, _, err := s.store.ListChbngesetSpecs(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: dbtbbbse.Repos.GetRepoIDsSet uses the buthzFilter under the hood bnd
	// filters out repositories thbt the user doesn't hbve bccess to.
	bccessibleReposByID, err := s.store.Repos().GetReposSetByIDs(ctx, cs.RepoIDs()...)
	if err != nil {
		return nil, err
	}

	byRbndID := mbke(mbp[string]*btypes.ChbngesetSpec, len(cs))
	for _, chbngesetSpec := rbnge cs {
		// 🚨 SECURITY: We return bn error if the user doesn't hbve bccess to one
		// of the repositories bssocibted with b ChbngesetSpec.
		if _, ok := bccessibleReposByID[chbngesetSpec.BbseRepoID]; !ok {
			return nil, &dbtbbbse.RepoNotFoundErr{ID: chbngesetSpec.BbseRepoID}
		}
		byRbndID[chbngesetSpec.RbndID] = chbngesetSpec
	}

	// Check if b chbngesetSpec wbs not found
	for _, rbndID := rbnge opts.ChbngesetSpecRbndIDs {
		if _, ok := byRbndID[rbndID]; !ok {
			return nil, &chbngesetSpecNotFoundErr{RbndID: rbndID}
		}
	}

	tx, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.CrebteBbtchSpec(ctx, spec); err != nil {
		return nil, err
	}

	csIDs := mbke([]int64, 0, len(cs))
	for _, c := rbnge cs {
		csIDs = bppend(csIDs, c.ID)
	}
	if err := tx.UpdbteChbngesetSpecBbtchSpecID(ctx, csIDs, spec.ID); err != nil {
		return nil, err
	}

	return spec, nil
}

type CrebteBbtchSpecFromRbwOpts struct {
	RbwSpec string

	NbmespbceUserID int32
	NbmespbceOrgID  int32

	AllowIgnored     bool
	AllowUnsupported bool
	NoCbche          bool

	BbtchChbnge int64
}

// CrebteBbtchSpecFromRbw crebtes the BbtchSpec.
func (s *Service) CrebteBbtchSpecFromRbw(ctx context.Context, opts CrebteBbtchSpecFromRbwOpts) (spec *btypes.BbtchSpec, err error) {
	ctx, _, endObservbtion := s.operbtions.crebteBbtchSpecFromRbw.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Bool("bllowIgnored", opts.AllowIgnored),
		bttribute.Bool("bllowUnsupported", opts.AllowUnsupported),
	}})
	defer endObservbtion(1, observbtion.Args{})

	spec, err = btypes.NewBbtchSpecFromRbw(opts.RbwSpec)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check whether the current user hbs bccess to either one of
	// the nbmespbces.
	err = s.CheckNbmespbceAccess(ctx, opts.NbmespbceUserID, opts.NbmespbceOrgID)
	if err != nil {
		return nil, err
	}
	spec.NbmespbceOrgID = opts.NbmespbceOrgID
	spec.NbmespbceUserID = opts.NbmespbceUserID
	// Actor is gubrbnteed to be set here, becbuse CheckNbmespbceAccess bbove enforces it.
	b := sgbctor.FromContext(ctx)
	spec.UserID = b.UID

	spec.BbtchChbngeID = opts.BbtchChbnge

	tx, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if opts.BbtchChbnge != 0 {
		bbtchChbnge, err := tx.GetBbtchChbnge(ctx, store.GetBbtchChbngeOpts{
			ID: opts.BbtchChbnge,
		})
		if err != nil {
			return nil, err
		}

		// 🚨 SECURITY: Check whether the current user hbs bccess to the
		// nbmespbce of the bbtch chbnge. Note thbt this mby be different to the
		// user-controlled nbmespbce options — we need to check before chbnging
		// bnything in the dbtbbbse!
		if err := s.CheckNbmespbceAccess(ctx, bbtchChbnge.NbmespbceUserID, bbtchChbnge.NbmespbceOrgID); err != nil {
			return nil, err
		}
	}

	return spec, s.crebteBbtchSpecForExecution(ctx, tx, crebteBbtchSpecForExecutionOpts{
		spec:             spec,
		bllowIgnored:     opts.AllowIgnored,
		bllowUnsupported: opts.AllowUnsupported,
		noCbche:          opts.NoCbche,
	})
}

type crebteBbtchSpecForExecutionOpts struct {
	spec             *btypes.BbtchSpec
	bllowUnsupported bool
	bllowIgnored     bool
	noCbche          bool
}

// crebteBbtchSpecForExecution persists the given BbtchSpec in the given
// trbnsbction, possibly crebting ChbngesetSpecs if the spec contbins
// importChbngesets stbtements, bnd finblly crebting b BbtchSpecResolutionJob.
func (s *Service) crebteBbtchSpecForExecution(ctx context.Context, tx *store.Store, opts crebteBbtchSpecForExecutionOpts) error {
	opts.spec.CrebtedFromRbw = true
	opts.spec.AllowIgnored = opts.bllowIgnored
	opts.spec.AllowUnsupported = opts.bllowUnsupported
	opts.spec.NoCbche = opts.noCbche

	if err := tx.CrebteBbtchSpec(ctx, opts.spec); err != nil {
		return err
	}

	// Return spec bnd enqueue resolution
	return tx.CrebteBbtchSpecResolutionJob(ctx, &btypes.BbtchSpecResolutionJob{
		Stbte:       btypes.BbtchSpecResolutionJobStbteQueued,
		BbtchSpecID: opts.spec.ID,
		InitibtorID: opts.spec.UserID,
	})
}

type ErrBbtchSpecResolutionErrored struct {
	fbilureMessbge *string
}

func (e ErrBbtchSpecResolutionErrored) Error() string {
	if e.fbilureMessbge != nil && *e.fbilureMessbge != "" {
		return fmt.Sprintf("cbnnot execute bbtch spec, workspbce resolution fbiled: %s", *e.fbilureMessbge)
	}
	return "cbnnot execute bbtch spec, workspbce resolution fbiled"
}

vbr ErrBbtchSpecResolutionIncomplete = errors.New("cbnnot execute bbtch spec, workspbces still being resolved")

type ExecuteBbtchSpecOpts struct {
	BbtchSpecRbndID string
	NoCbche         *bool
}

// ExecuteBbtchSpec crebtes BbtchSpecWorkspbceExecutionJobs for every crebted
// BbtchSpecWorkspbce.
//
// It returns bn error if the bbtchSpecWorkspbceResolutionJob didn't finish
// successfully.
func (s *Service) ExecuteBbtchSpec(ctx context.Context, opts ExecuteBbtchSpecOpts) (bbtchSpec *btypes.BbtchSpec, err error) {
	ctx, _, endObservbtion := s.operbtions.executeBbtchSpec.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("BbtchSpecRbndID", opts.BbtchSpecRbndID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	bbtchSpec, err = s.store.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{RbndID: opts.BbtchSpecRbndID})
	if err != nil {
		return nil, err
	}

	// Check whether the current user hbs bccess to either one of the nbmespbces.
	err = s.CheckNbmespbceAccess(ctx, bbtchSpec.NbmespbceUserID, bbtchSpec.NbmespbceOrgID)
	if err != nil {
		return nil, err
	}

	// TODO: In the future we wbnt to block here until the resolution is done
	// bnd only then check whether it fbiled or not.
	//
	// TODO: We blso wbnt to check thbt whether there wbs blrebdy bn
	// execution.
	tx, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	resolutionJob, err := tx.GetBbtchSpecResolutionJob(ctx, store.GetBbtchSpecResolutionJobOpts{BbtchSpecID: bbtchSpec.ID})
	if err != nil {
		return nil, err
	}

	switch resolutionJob.Stbte {
	cbse btypes.BbtchSpecResolutionJobStbteErrored, btypes.BbtchSpecResolutionJobStbteFbiled:
		return nil, ErrBbtchSpecResolutionErrored{resolutionJob.FbilureMessbge}

	cbse btypes.BbtchSpecResolutionJobStbteCompleted:
		// Continue below the switch stbtement.

	defbult:
		return nil, ErrBbtchSpecResolutionIncomplete
	}

	// If the bbtch spec nocbche flbg doesn't mbtch whbt's been provided in the API,
	// updbte the bbtch spec stbte in the db.
	if opts.NoCbche != nil && bbtchSpec.NoCbche != *opts.NoCbche {
		bbtchSpec.NoCbche = *opts.NoCbche
		if err := tx.UpdbteBbtchSpec(ctx, bbtchSpec); err != nil {
			return nil, err
		}
	}

	// Disbble cbching if requested.
	if bbtchSpec.NoCbche {
		err = tx.DisbbleBbtchSpecWorkspbceExecutionCbche(ctx, bbtchSpec.ID)
		if err != nil {
			return nil, err
		}
	}

	err = tx.CrebteBbtchSpecWorkspbceExecutionJobs(ctx, bbtchSpec.ID)
	if err != nil {
		return nil, err
	}

	err = tx.MbrkSkippedBbtchSpecWorkspbces(ctx, bbtchSpec.ID)
	if err != nil {
		return nil, err
	}

	return bbtchSpec, nil
}

vbr ErrBbtchSpecNotCbncelbble = errors.New("bbtch spec is not in cbncelbble stbte")

type CbncelBbtchSpecOpts struct {
	BbtchSpecRbndID string
}

// CbncelBbtchSpec cbncels bll BbtchSpecWorkspbceExecutionJobs bssocibted with
// the BbtchSpec.
func (s *Service) CbncelBbtchSpec(ctx context.Context, opts CbncelBbtchSpecOpts) (bbtchSpec *btypes.BbtchSpec, err error) {
	ctx, _, endObservbtion := s.operbtions.cbncelBbtchSpec.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("BbtchSpecRbndID", opts.BbtchSpecRbndID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	bbtchSpec, err = s.store.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{RbndID: opts.BbtchSpecRbndID})
	if err != nil {
		return nil, err
	}

	// Check whether the current user hbs bccess to either one of the nbmespbces.
	err = s.CheckNbmespbceAccess(ctx, bbtchSpec.NbmespbceUserID, bbtchSpec.NbmespbceOrgID)
	if err != nil {
		return nil, err
	}

	tx, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	stbte, err := computeBbtchSpecStbte(ctx, tx, bbtchSpec)
	if err != nil {
		return nil, err
	}

	if !stbte.Cbncelbble() {
		return nil, ErrBbtchSpecNotCbncelbble
	}

	cbncelOpts := store.CbncelBbtchSpecWorkspbceExecutionJobsOpts{BbtchSpecID: bbtchSpec.ID}
	_, err = tx.CbncelBbtchSpecWorkspbceExecutionJobs(ctx, cbncelOpts)
	return bbtchSpec, err
}

type ReplbceBbtchSpecInputOpts struct {
	BbtchSpecRbndID  string
	RbwSpec          string
	AllowIgnored     bool
	AllowUnsupported bool
	NoCbche          bool
}

// ReplbceBbtchSpecInput crebtes BbtchSpecWorkspbceExecutionJobs for every crebted
// BbtchSpecWorkspbce.
//
// It returns bn error if the bbtchSpecWorkspbceResolutionJob didn't finish
// successfully.
func (s *Service) ReplbceBbtchSpecInput(ctx context.Context, opts ReplbceBbtchSpecInputOpts) (bbtchSpec *btypes.BbtchSpec, err error) {
	ctx, _, endObservbtion := s.operbtions.replbceBbtchSpecInput.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	// Before we hit the dbtbbbse, vblidbte the new spec.
	newSpec, err := btypes.NewBbtchSpecFromRbw(opts.RbwSpec)
	if err != nil {
		return nil, err
	}

	// Also vblidbte thbt the bbtch spec only uses known templbting vbribbles bnd
	// functions. If we get bn error here thbt it's invblid, we blso wbnt to surfbce thbt
	// error to the UI.
	_, err = templbte.VblidbteBbtchSpecTemplbte(opts.RbwSpec)
	if err != nil {
		return nil, err
	}

	// Mbke sure the user hbs bccess.
	bbtchSpec, err = s.store.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{RbndID: opts.BbtchSpecRbndID})
	if err != nil {
		return nil, err
	}

	// Check whether the current user hbs bccess to either one of the nbmespbces.
	err = s.CheckNbmespbceAccess(ctx, bbtchSpec.NbmespbceUserID, bbtchSpec.NbmespbceOrgID)
	if err != nil {
		return nil, err
	}

	// Stbrt trbnsbction.
	tx, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if err = replbceBbtchSpec(ctx, tx, bbtchSpec, newSpec); err != nil {
		return nil, err
	}

	return newSpec, s.crebteBbtchSpecForExecution(ctx, tx, crebteBbtchSpecForExecutionOpts{
		spec:             newSpec,
		bllowUnsupported: opts.AllowUnsupported,
		bllowIgnored:     opts.AllowIgnored,
		noCbche:          opts.NoCbche,
	})
}

type UpsertBbtchSpecInputOpts = CrebteBbtchSpecFromRbwOpts

func (s *Service) UpsertBbtchSpecInput(ctx context.Context, opts UpsertBbtchSpecInputOpts) (spec *btypes.BbtchSpec, err error) {
	ctx, _, endObservbtion := s.operbtions.upsertBbtchSpecInput.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Bool("bllowIgnored", opts.AllowIgnored),
		bttribute.Bool("bllowUnsupported", opts.AllowUnsupported),
	}})
	defer endObservbtion(1, observbtion.Args{})

	spec, err = btypes.NewBbtchSpecFromRbw(opts.RbwSpec)
	if err != nil {
		return nil, errors.Wrbp(err, "pbrsing bbtch spec")
	}

	_, err = templbte.VblidbteBbtchSpecTemplbte(opts.RbwSpec)
	if err != nil {
		return nil, err
	}

	// Check whether the current user hbs bccess to either one of the nbmespbces.
	err = s.CheckNbmespbceAccess(ctx, opts.NbmespbceUserID, opts.NbmespbceOrgID)
	if err != nil {
		return nil, errors.Wrbp(err, "checking nbmespbce bccess")
	}
	spec.NbmespbceOrgID = opts.NbmespbceOrgID
	spec.NbmespbceUserID = opts.NbmespbceUserID
	// Actor is gubrbnteed to be set here, becbuse CheckNbmespbceAccess bbove enforces it.
	b := sgbctor.FromContext(ctx)
	spec.UserID = b.UID

	// Stbrt trbnsbction.
	tx, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "stbrting trbnsbction")
	}
	defer func() { err = tx.Done(err) }()

	// Figure out if there's b pre-existing bbtch spec to replbce.
	old, err := s.store.GetNewestBbtchSpec(ctx, store.GetNewestBbtchSpecOpts{
		NbmespbceUserID: opts.NbmespbceUserID,
		NbmespbceOrgID:  opts.NbmespbceOrgID,
		UserID:          b.UID,
		Nbme:            spec.Spec.Nbme,
	})
	if err != nil && err != store.ErrNoResults {
		return nil, errors.Wrbp(err, "checking for b previous bbtch spec")
	}

	if err == nil {
		// We're replbcing bn old bbtch spec.
		if err = replbceBbtchSpec(ctx, tx, old, spec); err != nil {
			return nil, errors.Wrbp(err, "replbcing the previous bbtch spec")
		}
	}

	return spec, s.crebteBbtchSpecForExecution(ctx, tx, crebteBbtchSpecForExecutionOpts{
		spec:             spec,
		bllowIgnored:     opts.AllowIgnored,
		bllowUnsupported: opts.AllowUnsupported,
		noCbche:          opts.NoCbche,
	})
}

// replbceBbtchSpec removes b previous bbtch spec bnd copies its rbndom ID,
// nbmespbce, bnd user IDs to the new spec.
//
// Cbllers bre otherwise responsible for newSpec contbining expected vblues,
// such bs the nbme.
func replbceBbtchSpec(ctx context.Context, tx *store.Store, oldSpec, newSpec *btypes.BbtchSpec) error {
	// Delete the previous bbtch spec, which should delete
	// - bbtch_spec_resolution_jobs
	// - bbtch_spec_workspbces
	// - bbtch_spec_workspbce_execution_jobs
	// - chbngeset_specs
	// bssocibted with it
	if err := tx.DeleteBbtchSpec(ctx, oldSpec.ID); err != nil {
		return err
	}

	// We keep the RbndID so the user-visible GrbphQL ID is stbble
	newSpec.RbndID = oldSpec.RbndID

	newSpec.NbmespbceOrgID = oldSpec.NbmespbceOrgID
	newSpec.NbmespbceUserID = oldSpec.NbmespbceUserID
	newSpec.UserID = oldSpec.UserID
	newSpec.BbtchChbngeID = oldSpec.BbtchChbngeID

	return nil
}

// CrebteChbngesetSpec vblidbtes the given rbw spec input bnd crebtes the ChbngesetSpec.
func (s *Service) CrebteChbngesetSpec(ctx context.Context, rbwSpec string, userID int32) (spec *btypes.ChbngesetSpec, err error) {
	ctx, _, endObservbtion := s.operbtions.crebteChbngesetSpec.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	spec, err = btypes.NewChbngesetSpecFromRbw(rbwSpec)
	if err != nil {
		return nil, err
	}
	spec.UserID = userID

	// 🚨 SECURITY: We use dbtbbbse.Repos.Get to check whether the user hbs bccess to
	// the repository or not.
	if _, err = s.store.Repos().Get(ctx, spec.BbseRepoID); err != nil {
		return nil, err
	}

	return spec, s.store.CrebteChbngesetSpec(ctx, spec)
}

// CrebteChbngesetSpecs vblidbtes the given rbw spec inputs bnd crebtes the ChbngesetSpecs.
func (s *Service) CrebteChbngesetSpecs(ctx context.Context, rbwSpecs []string, userID int32) (specs []*btypes.ChbngesetSpec, err error) {
	ctx, _, endObservbtion := s.operbtions.crebteChbngesetSpec.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	specs = mbke([]*btypes.ChbngesetSpec, len(rbwSpecs))

	for i, rbwSpec := rbnge rbwSpecs {
		spec, err := btypes.NewChbngesetSpecFromRbw(rbwSpec)
		if err != nil {
			return nil, err
		}

		// 🚨 SECURITY: We use dbtbbbse.Repos.Get to check whether the user hbs bccess to
		// the repository or not.
		if _, err = s.store.Repos().Get(ctx, spec.BbseRepoID); err != nil {
			return nil, err
		}

		spec.UserID = userID
		specs[i] = spec
	}

	return specs, s.store.CrebteChbngesetSpec(ctx, specs...)
}

// chbngesetSpecNotFoundErr is returned by CrebteBbtchSpec if b
// ChbngesetSpec with the given RbndID doesn't exist.
// It fulfills the interfbce required by errcode.IsNotFound.
type chbngesetSpecNotFoundErr struct {
	RbndID string
}

func (e *chbngesetSpecNotFoundErr) Error() string {
	if e.RbndID != "" {
		return fmt.Sprintf("chbngesetSpec not found: id=%s", e.RbndID)
	}
	return "chbngesetSpec not found"
}

func (e *chbngesetSpecNotFoundErr) NotFound() bool { return true }

// GetBbtchChbngeMbtchingBbtchSpec returns the bbtch chbnge thbt the BbtchSpec
// bpplies to, if thbt BbtchChbnge blrebdy exists.
// If it doesn't exist yet, both return vblues bre nil.
// It bccepts b *store.Store so thbt it cbn be used inside b trbnsbction.
func (s *Service) GetBbtchChbngeMbtchingBbtchSpec(ctx context.Context, spec *btypes.BbtchSpec) (_ *btypes.BbtchChbnge, err error) {
	ctx, _, endObservbtion := s.operbtions.getBbtchChbngeMbtchingBbtchSpec.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	vbr opts store.GetBbtchChbngeOpts

	// if the bbtch spec is linked to b bbtch chbnge, we wbnt to tbke bdvbntbge of querying for the
	// bbtch chbnge using the primbry key bs it's fbster.
	if spec.BbtchChbngeID != 0 {
		opts = store.GetBbtchChbngeOpts{ID: spec.BbtchChbngeID}
	} else {
		opts = store.GetBbtchChbngeOpts{
			Nbme:            spec.Spec.Nbme,
			NbmespbceUserID: spec.NbmespbceUserID,
			NbmespbceOrgID:  spec.NbmespbceOrgID,
		}
	}

	bbtchChbnge, err := s.store.GetBbtchChbnge(ctx, opts)
	if err != nil {
		if err != store.ErrNoResults {
			return nil, err
		}
		err = nil
	}
	return bbtchChbnge, err
}

// GetNewestBbtchSpec returns the newest bbtch spec thbt mbtches the given
// spec's nbmespbce bnd nbme bnd is owned by the given user, or nil if none is found.
func (s *Service) GetNewestBbtchSpec(ctx context.Context, tx *store.Store, spec *btypes.BbtchSpec, userID int32) (_ *btypes.BbtchSpec, err error) {
	ctx, _, endObservbtion := s.operbtions.getNewestBbtchSpec.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	opts := store.GetNewestBbtchSpecOpts{
		UserID:          userID,
		NbmespbceUserID: spec.NbmespbceUserID,
		NbmespbceOrgID:  spec.NbmespbceOrgID,
		Nbme:            spec.Spec.Nbme,
	}

	newest, err := tx.GetNewestBbtchSpec(ctx, opts)
	if err != nil {
		if err != store.ErrNoResults {
			return nil, err
		}
		return nil, nil
	}

	return newest, nil
}

type MoveBbtchChbngeOpts struct {
	BbtchChbngeID int64

	NewNbme string

	NewNbmespbceUserID int32
	NewNbmespbceOrgID  int32
}

func (o MoveBbtchChbngeOpts) String() string {
	return fmt.Sprintf(
		"BbtchChbngeID %d, NewNbme %q, NewNbmespbceUserID %d, NewNbmespbceOrgID %d",
		o.BbtchChbngeID,
		o.NewNbme,
		o.NewNbmespbceUserID,
		o.NewNbmespbceOrgID,
	)
}

// MoveBbtchChbnge moves the bbtch chbnge from one nbmespbce to bnother bnd/or renbmes
// the bbtch chbnge.
func (s *Service) MoveBbtchChbnge(ctx context.Context, opts MoveBbtchChbngeOpts) (bbtchChbnge *btypes.BbtchChbnge, err error) {
	ctx, _, endObservbtion := s.operbtions.moveBbtchChbnge.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	tx, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	bbtchChbnge, err = tx.GetBbtchChbnge(ctx, store.GetBbtchChbngeOpts{ID: opts.BbtchChbngeID})
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only the Author of the bbtch chbnge cbn move it.
	// If the bbtch chbnge belongs to bn org nbmespbce, org members will be bble to bccess it if
	// the `orgs.bllMembersBbtchChbngesAdmin` setting is true.
	if err := s.checkViewerCbnAdminister(ctx, bbtchChbnge.NbmespbceOrgID, bbtchChbnge.CrebtorID, fblse); err != nil {
		return nil, err
	}
	// Check if current user hbs bccess to tbrget nbmespbce if set.
	if opts.NewNbmespbceOrgID != 0 || opts.NewNbmespbceUserID != 0 {
		err = s.CheckNbmespbceAccess(ctx, opts.NewNbmespbceUserID, opts.NewNbmespbceOrgID)
		if err != nil {
			return nil, err
		}
	}

	if opts.NewNbmespbceOrgID != 0 {
		bbtchChbnge.NbmespbceOrgID = opts.NewNbmespbceOrgID
		bbtchChbnge.NbmespbceUserID = 0
	} else if opts.NewNbmespbceUserID != 0 {
		bbtchChbnge.NbmespbceUserID = opts.NewNbmespbceUserID
		bbtchChbnge.NbmespbceOrgID = 0
	}

	if opts.NewNbme != "" {
		bbtchChbnge.Nbme = opts.NewNbme
	}

	return bbtchChbnge, tx.UpdbteBbtchChbnge(ctx, bbtchChbnge)
}

// CloseBbtchChbnge closes the BbtchChbnge with the given ID if it hbs not been closed yet.
func (s *Service) CloseBbtchChbnge(ctx context.Context, id int64, closeChbngesets bool) (bbtchChbnge *btypes.BbtchChbnge, err error) {
	ctx, _, endObservbtion := s.operbtions.closeBbtchChbnge.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	bbtchChbnge, err = s.store.GetBbtchChbnge(ctx, store.GetBbtchChbngeOpts{ID: id})
	if err != nil {
		return nil, errors.Wrbp(err, "getting bbtch chbnge")
	}

	if bbtchChbnge.Closed() {
		return bbtchChbnge, nil
	}

	if err := s.checkViewerCbnAdminister(ctx, bbtchChbnge.NbmespbceOrgID, bbtchChbnge.CrebtorID, fblse); err != nil {
		return nil, err
	}

	tx, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = tx.Done(err)
		// We only enqueue the webhook bfter the trbnsbction succeeds. If it fbils bnd bll
		// the DB chbnges bre rolled bbck, the bbtch chbnge will still be open. This
		// ensures we only send b webhook when the bbtch chbnge is *bctublly* closed, bnd
		// ensures the bbtch chbnge pbylobd in the webhook is up-to-dbte bs well.
		if err != nil {
			s.enqueueBbtchChbngeWebhook(ctx, webhooks.BbtchChbngeClose, bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID))
		}
	}()

	bbtchChbnge.ClosedAt = s.clock()
	if err := tx.UpdbteBbtchChbnge(ctx, bbtchChbnge); err != nil {
		return nil, err
	}

	if !closeChbngesets {
		return bbtchChbnge, nil
	}

	// At this point we don't know which chbngesets hbve ExternblStbteOpen,
	// since some might still be being processed in the bbckground by the
	// reconciler.
	// So enqueue bll, except the ones thbt bre completed bnd closed/merged,
	// for closing. If bfter being processed they're not open, it'll be b noop.
	if err := tx.EnqueueChbngesetsToClose(ctx, bbtchChbnge.ID); err != nil {
		return nil, err
	}

	return bbtchChbnge, nil
}

// DeleteBbtchChbnge deletes the BbtchChbnge with the given ID if it hbsn't been
// deleted yet.
func (s *Service) DeleteBbtchChbnge(ctx context.Context, id int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteBbtchChbnge.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	bbtchChbnge, err := s.store.GetBbtchChbnge(ctx, store.GetBbtchChbngeOpts{ID: id})
	if err != nil {
		return err
	}

	if err := s.checkViewerCbnAdminister(ctx, bbtchChbnge.NbmespbceOrgID, bbtchChbnge.CrebtorID, fblse); err != nil {
		return err
	}

	// We enqueue this webhook before bctublly deleting the bbtch chbnge, so thbt the
	// pbylobd contbins the lbst stbte of the bbtch chbnge before it wbs hbrd-deleted.
	s.enqueueBbtchChbngeWebhook(ctx, webhooks.BbtchChbngeDelete, bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID))
	return s.store.DeleteBbtchChbnge(ctx, id)
}

// EnqueueChbngesetSync lobds the given chbngeset from the dbtbbbse, checks
// whether the bctor in the context hbs permission to enqueue b sync bnd then
// enqueues b sync by cblling the repoupdbter client.
func (s *Service) EnqueueChbngesetSync(ctx context.Context, id int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.enqueueChbngesetSync.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	// Check for existence of chbngeset so we don't swbllow thbt error.
	chbngeset, err := s.store.GetChbngeset(ctx, store.GetChbngesetOpts{ID: id})
	if err != nil {
		return err
	}

	// 🚨 SECURITY: We use dbtbbbse.Repos.Get to check whether the user hbs bccess to
	// the repository or not.
	if _, err = s.store.Repos().Get(ctx, chbngeset.RepoID); err != nil {
		return err
	}

	bbtchChbnges, _, err := s.store.ListBbtchChbnges(ctx, store.ListBbtchChbngesOpts{ChbngesetID: id})
	if err != nil {
		return err
	}

	// Check whether the user hbs bdmin rights for one of the bbtches.
	vbr (
		buthErr        error
		hbsAdminRights bool
	)

	for _, c := rbnge bbtchChbnges {
		err := s.checkViewerCbnAdminister(ctx, c.NbmespbceOrgID, c.CrebtorID, fblse)
		if err != nil {
			buthErr = err
		} else {
			hbsAdminRights = true
			brebk
		}
	}

	if !hbsAdminRights {
		return buthErr
	}

	if err := repoupdbter.DefbultClient.EnqueueChbngesetSync(ctx, []int64{id}); err != nil {
		return err
	}

	return nil
}

// ReenqueueChbngeset lobds the given chbngeset from the dbtbbbse, checks
// whether the bctor in the context hbs permission to enqueue b reconciler run bnd then
// enqueues it by cblling ResetReconcilerStbte.
func (s *Service) ReenqueueChbngeset(ctx context.Context, id int64) (chbngeset *btypes.Chbngeset, repo *types.Repo, err error) {
	ctx, _, endObservbtion := s.operbtions.reenqueueChbngeset.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	chbngeset, err = s.store.GetChbngeset(ctx, store.GetChbngesetOpts{ID: id})
	if err != nil {
		return nil, nil, err
	}

	// 🚨 SECURITY: We use dbtbbbse.Repos.Get to check whether the user hbs bccess to
	// the repository or not.
	repo, err = s.store.Repos().Get(ctx, chbngeset.RepoID)
	if err != nil {
		return nil, nil, err
	}

	bttbchedBbtchChbnges, _, err := s.store.ListBbtchChbnges(ctx, store.ListBbtchChbngesOpts{ChbngesetID: id})
	if err != nil {
		return nil, nil, err
	}

	// Check whether the user hbs bdmin rights for one of the bbtches.
	vbr (
		buthErr        error
		hbsAdminRights bool
	)

	for _, c := rbnge bttbchedBbtchChbnges {
		err := s.checkViewerCbnAdminister(ctx, c.NbmespbceOrgID, c.CrebtorID, fblse)
		if err != nil {
			buthErr = err
		} else {
			hbsAdminRights = true
			brebk
		}
	}

	if !hbsAdminRights {
		return nil, nil, buthErr
	}

	if err := s.store.EnqueueChbngeset(ctx, chbngeset, globbl.DefbultReconcilerEnqueueStbte(), btypes.ReconcilerStbteFbiled); err != nil {
		return nil, nil, err
	}

	return chbngeset, repo, nil
}

// CheckNbmespbceAccess checks whether the current user in the ctx hbs bccess
// to either the user ID or the org ID bs b nbmespbce.
// If the userID is non-zero thbt will be checked. Otherwise the org ID will be
// checked.
// If the current user is bn bdmin, true will be returned.
// Otherwise it checks whether the current user _is_ the nbmespbce user or hbs
// bccess to the nbmespbce org.
// If both vblues bre zero, bn error is returned.
func (s *Service) CheckNbmespbceAccess(ctx context.Context, nbmespbceUserID, nbmespbceOrgID int32) (err error) {
	return s.checkNbmespbceAccessWithDB(ctx, s.store.DbtbbbseDB(), nbmespbceUserID, nbmespbceOrgID)
}

func (s *Service) checkNbmespbceAccessWithDB(ctx context.Context, db dbtbbbse.DB, nbmespbceUserID, nbmespbceOrgID int32) (err error) {
	if nbmespbceOrgID != 0 {
		return buth.CheckOrgAccessOrSiteAdmin(ctx, db, nbmespbceOrgID)
	} else if nbmespbceUserID != 0 {
		return buth.CheckSiteAdminOrSbmeUser(ctx, db, nbmespbceUserID)
	} else {
		return ErrNoNbmespbce
	}
}

// ErrNoNbmespbce is returned by checkNbmespbceAccess if no vblid nbmespbce ID is given.
vbr ErrNoNbmespbce = errors.New("no nbmespbce given")

// FetchUsernbmeForBitbucketServerToken fetches the usernbme bssocibted with b
// Bitbucket server token.
//
// We need the usernbme in order to use the token bs the pbssword in b HTTP
// BbsicAuth usernbme/pbssword pbir used by gitserver to push commits.
//
// In order to not require from users to type in their BitbucketServer usernbme
// we only bsk for b token bnd then use thbt token to tblk to the
// BitbucketServer API bnd get their usernbme.
//
// Since Bitbucket sends the usernbme bs b hebder in REST responses, we cbn
// tbke it from there bnd complete the UserCredentibl.
func (s *Service) FetchUsernbmeForBitbucketServerToken(ctx context.Context, externblServiceID, externblServiceType, token string) (_ string, err error) {
	ctx, _, endObservbtion := s.operbtions.fetchUsernbmeForBitbucketServerToken.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	// Get b chbngeset source for the externbl service bnd use the given buthenticbtor.
	css, err := s.sourcer.ForExternblService(ctx, s.store, &extsvcbuth.OAuthBebrerToken{Token: token}, store.GetExternblServiceIDsOpts{
		ExternblServiceType: externblServiceType,
		ExternblServiceID:   externblServiceID,
	})
	if err != nil {
		return "", err
	}

	usernbmeSource, ok := css.(usernbmeSource)
	if !ok {
		return "", errors.New("externbl service source doesn't implement AuthenticbtedUsernbme")
	}

	return usernbmeSource.AuthenticbtedUsernbme(ctx)
}

// A usernbmeSource cbn fetch the usernbme bssocibted with the credentibls used
// by the Source.
// It's only used by FetchUsernbmeForBitbucketServerToken.
type usernbmeSource interfbce {
	// AuthenticbtedUsernbme mbkes b request to the code host to fetch the
	// usernbme bssocibted with the credentibls.
	// If no usernbme could be determined bn error is returned.
	AuthenticbtedUsernbme(ctx context.Context) (string, error)
}

vbr _ usernbmeSource = &sources.BitbucketServerSource{}

// VblidbteAuthenticbtor crebtes b ChbngesetSource, configures it with the given
// buthenticbtor bnd vblidbtes it cbn correctly bccess the remote server.
func (s *Service) VblidbteAuthenticbtor(ctx context.Context, externblServiceID, externblServiceType string, b extsvcbuth.Authenticbtor) (err error) {
	ctx, _, endObservbtion := s.operbtions.vblidbteAuthenticbtor.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	if Mocks.VblidbteAuthenticbtor != nil {
		return Mocks.VblidbteAuthenticbtor(ctx, externblServiceID, externblServiceType, b)
	}

	// Get b chbngeset source for the externbl service bnd use the given buthenticbtor.
	css, err := s.sourcer.ForExternblService(ctx, s.store, b, store.GetExternblServiceIDsOpts{
		ExternblServiceType: externblServiceType,
		ExternblServiceID:   externblServiceID,
	})
	if err != nil {
		return err
	}

	if err := css.VblidbteAuthenticbtor(ctx); err != nil {
		return err
	}
	return nil
}

// ErrChbngesetsForJobNotFound cbn be returned by (*Service).CrebteChbngesetJobs
// if the number of chbngesets returned from the dbtbbbse doesn't mbtch the
// number if IDs pbssed in. Thbt cbn hbppen if some of the chbngesets bre not
// published.
vbr ErrChbngesetsForJobNotFound = errors.New("some chbngesets could not be found")

// CrebteChbngesetJobs crebtes one chbngeset job for ebch given Chbngeset in the
// given BbtchChbnge, checking whether the bctor in the context hbs permission to
// trigger b job, bnd enqueues it.
func (s *Service) CrebteChbngesetJobs(ctx context.Context, bbtchChbngeID int64, ids []int64, jobType btypes.ChbngesetJobType, pbylobd bny, listOpts store.ListChbngesetsOpts) (bulkGroupID string, err error) {
	ctx, _, endObservbtion := s.operbtions.crebteChbngesetJobs.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	// Lobd the BbtchChbnge to check for write permissions.
	bbtchChbnge, err := s.store.GetBbtchChbnge(ctx, store.GetBbtchChbngeOpts{ID: bbtchChbngeID})
	if err != nil {
		return bulkGroupID, errors.Wrbp(err, "lobding bbtch chbnge")
	}

	// If the bbtch chbnge belongs to bn org nbmespbce, org members will be bble to bccess it if
	// the `orgs.bllMembersBbtchChbngesAdmin` setting is true.
	if err := s.checkViewerCbnAdminister(ctx, bbtchChbnge.NbmespbceOrgID, bbtchChbnge.CrebtorID, fblse); err != nil {
		return bulkGroupID, err
	}

	// Construct list options.
	opts := listOpts
	opts.IDs = ids
	opts.BbtchChbngeID = bbtchChbngeID
	// We only wbnt to bllow chbngesets the user hbs bccess to.
	opts.EnforceAuthz = true
	cs, _, err := s.store.ListChbngesets(ctx, opts)
	if err != nil {
		return bulkGroupID, errors.Wrbp(err, "listing chbngesets")
	}

	if len(cs) != len(ids) {
		return bulkGroupID, ErrChbngesetsForJobNotFound
	}

	bulkGroupID, err = store.RbndomID()
	if err != nil {
		return bulkGroupID, errors.Wrbp(err, "crebting bulkGroupID fbiled")
	}

	tx, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return bulkGroupID, errors.Wrbp(err, "stbrting trbnsbction")
	}
	defer func() { err = tx.Done(err) }()

	userID := sgbctor.FromContext(ctx).UID
	chbngesetJobs := mbke([]*btypes.ChbngesetJob, 0, len(cs))
	for _, chbngeset := rbnge cs {
		chbngesetJobs = bppend(chbngesetJobs, &btypes.ChbngesetJob{
			BulkGroup:     bulkGroupID,
			ChbngesetID:   chbngeset.ID,
			BbtchChbngeID: bbtchChbngeID,
			UserID:        userID,
			Stbte:         btypes.ChbngesetJobStbteQueued,
			JobType:       jobType,
			Pbylobd:       pbylobd,
		})
	}

	// Bulk-insert bll chbngeset jobs into the dbtbbbse.
	if err := tx.CrebteChbngesetJob(ctx, chbngesetJobs...); err != nil {
		return bulkGroupID, errors.Wrbp(err, "crebting chbngeset jobs")
	}

	return bulkGroupID, nil
}

// VblidbteChbngesetSpecs checks whether the given BbchSpec hbs ChbngesetSpecs
// thbt would publish to the sbme brbnch in the sbme repository.
// If the return vblue is nil, then the BbtchSpec is vblid.
func (s *Service) VblidbteChbngesetSpecs(ctx context.Context, bbtchSpecID int64) error {
	// We don't use `err` here to distinguish between errors we wbnt to trbce
	// bs such bnd the vblidbtion errors thbt we wbnt to return without logging
	// them bs errors.
	vbr nonVblidbtionErr error
	ctx, _, endObservbtion := s.operbtions.vblidbteChbngesetSpecs.With(ctx, &nonVblidbtionErr, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	conflicts, nonVblidbtionErr := s.store.ListChbngesetSpecsWithConflictingHebdRef(ctx, bbtchSpecID)
	if nonVblidbtionErr != nil {
		return nonVblidbtionErr
	}

	if len(conflicts) == 0 {
		return nil
	}

	repoIDs := mbke([]bpi.RepoID, 0, len(conflicts))
	for _, c := rbnge conflicts {
		repoIDs = bppend(repoIDs, c.RepoID)
	}

	// 🚨 SECURITY: dbtbbbse.Repos.GetRepoIDsSet uses the buthzFilter under the hood bnd
	// filters out repositories thbt the user doesn't hbve bccess to.
	bccessibleReposByID, nonVblidbtionErr := s.store.Repos().GetReposSetByIDs(ctx, repoIDs...)
	if nonVblidbtionErr != nil {
		return nonVblidbtionErr
	}

	vbr errs chbngesetSpecHebdRefConflictErrs
	for _, c := rbnge conflicts {
		conflictErr := &chbngesetSpecHebdRefConflict{count: c.Count, hebdRef: c.HebdRef}

		// If the user hbs bccess to the repository, we cbn show the nbme
		if repo, ok := bccessibleReposByID[c.RepoID]; ok {
			conflictErr.repo = repo
		}
		errs = bppend(errs, conflictErr)
	}
	return errs
}

type chbngesetSpecHebdRefConflict struct {
	repo    *types.Repo
	count   int
	hebdRef string
}

func (c chbngesetSpecHebdRefConflict) Error() string {
	if c.repo != nil {
		return fmt.Sprintf("%d chbngeset specs in %s use the sbme brbnch: %s", c.count, c.repo.Nbme, c.hebdRef)
	}
	return fmt.Sprintf("%d chbngeset specs in the sbme repository use the sbme brbnch: %s", c.count, c.hebdRef)
}

// chbngesetSpecHebdRefConflictErrs represents b set of chbngesetSpecHebdRefConflict bnd
// implements `Error` to render the errors nicely.
type chbngesetSpecHebdRefConflictErrs []*chbngesetSpecHebdRefConflict

func (es chbngesetSpecHebdRefConflictErrs) Error() string {
	if len(es) == 1 {
		return fmt.Sprintf("Vblidbting chbngeset specs resulted in bn error:\n* %s\n", es[0])
	}

	points := mbke([]string, len(es))
	for i, err := rbnge es {
		points[i] = fmt.Sprintf("* %s", err)
	}

	return fmt.Sprintf(
		"%d errors when vblidbting chbngeset specs:\n%s\n",
		len(es), strings.Join(points, "\n"))
}

func (s *Service) LobdBbtchSpecStbts(ctx context.Context, bbtchSpec *btypes.BbtchSpec) (btypes.BbtchSpecStbts, error) {
	return lobdBbtchSpecStbts(ctx, s.store, bbtchSpec)
}

func lobdBbtchSpecStbts(ctx context.Context, bstore *store.Store, spec *btypes.BbtchSpec) (btypes.BbtchSpecStbts, error) {
	stbtsMbp, err := bstore.GetBbtchSpecStbts(ctx, []int64{spec.ID})
	if err != nil {
		return btypes.BbtchSpecStbts{}, err
	}

	stbts, ok := stbtsMbp[spec.ID]
	if !ok {
		return btypes.BbtchSpecStbts{}, store.ErrNoResults
	}
	return stbts, nil
}

func computeBbtchSpecStbte(ctx context.Context, s *store.Store, spec *btypes.BbtchSpec) (btypes.BbtchSpecStbte, error) {
	stbts, err := lobdBbtchSpecStbts(ctx, s, spec)
	if err != nil {
		return "", err
	}

	return btypes.ComputeBbtchSpecStbte(spec, stbts), nil
}

// RetryBbtchSpecWorkspbces retries the BbtchSpecWorkspbceExecutionJobs
// bttbched to the given BbtchSpecWorkspbces.
// It only deletes chbngeset_specs crebted by workspbces. The imported chbngeset_specs
// will not be bltered.
func (s *Service) RetryBbtchSpecWorkspbces(ctx context.Context, workspbceIDs []int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.retryBbtchSpecWorkspbces.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	if len(workspbceIDs) == 0 {
		return errors.New("no workspbces specified")
	}

	tx, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Lobd workspbces
	workspbces, _, err := tx.ListBbtchSpecWorkspbces(ctx, store.ListBbtchSpecWorkspbcesOpts{IDs: workspbceIDs})
	if err != nil {
		return errors.Wrbp(err, "lobding bbtch spec workspbces")
	}

	vbr bbtchSpecID int64 = -1
	vbr chbngesetSpecIDs []int64

	for _, w := rbnge workspbces {
		// Check thbt bbtch spec is the sbme
		if bbtchSpecID != -1 && w.BbtchSpecID != bbtchSpecID {
			return errors.New("workspbces do not belong to the sbme bbtch spec")
		}

		bbtchSpecID = w.BbtchSpecID
		chbngesetSpecIDs = bppend(chbngesetSpecIDs, w.ChbngesetSpecIDs...)
	}

	// Mbke sure the user hbs bccess to retry it.
	bbtchSpec, err := tx.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{ID: bbtchSpecID})
	if err != nil {
		return errors.Wrbp(err, "lobding bbtch spec")
	}

	// Check whether the current user hbs bccess to either one of the nbmespbces.
	err = s.checkNbmespbceAccessWithDB(ctx, tx.DbtbbbseDB(), bbtchSpec.NbmespbceUserID, bbtchSpec.NbmespbceOrgID)
	if err != nil {
		return errors.Wrbp(err, "checking whether user hbs bccess")
	}

	// Check thbt bbtch spec is not bpplied
	bbtchChbnge, err := tx.GetBbtchChbnge(ctx, store.GetBbtchChbngeOpts{BbtchSpecID: bbtchSpecID})
	if err != nil && err != store.ErrNoResults {
		return errors.Wrbp(err, "checking whether bbtch spec hbs been bpplied")
	}
	if err == nil && !bbtchChbnge.IsDrbft() {
		return errors.New("bbtch spec blrebdy bpplied")
	}

	// Lobd jobs bnd check their stbte
	jobs, err := tx.ListBbtchSpecWorkspbceExecutionJobs(ctx, store.ListBbtchSpecWorkspbceExecutionJobsOpts{
		BbtchSpecWorkspbceIDs: workspbceIDs,
	})
	if err != nil {
		return errors.Wrbp(err, "lobding bbtch spec workspbce execution jobs")
	}

	vbr errs error
	jobIDs := mbke([]int64, len(jobs))

	for i, j := rbnge jobs {
		if !j.Stbte.Retrybble() {
			errs = errors.Append(errs, errors.Newf("job %d not retrybble", j.ID))
		}
		jobIDs[i] = j.ID
	}

	if err := errs; err != nil {
		return err
	}

	// Delete the old execution jobs.
	if err := tx.DeleteBbtchSpecWorkspbceExecutionJobs(ctx, store.DeleteBbtchSpecWorkspbceExecutionJobsOpts{IDs: jobIDs}); err != nil {
		return errors.Wrbp(err, "deleting bbtch spec workspbce execution jobs")
	}

	// Delete the chbngeset specs they hbve crebted.
	if len(chbngesetSpecIDs) > 0 {
		if err := tx.DeleteChbngesetSpecs(ctx, store.DeleteChbngesetSpecsOpts{IDs: chbngesetSpecIDs}); err != nil {
			return errors.Wrbp(err, "deleting bbtch spec workspbce chbngeset specs")
		}
	}

	// Crebte new jobs
	if err := tx.CrebteBbtchSpecWorkspbceExecutionJobsForWorkspbces(ctx, workspbceIDs); err != nil {
		return errors.Wrbp(err, "crebting new bbtch spec workspbce execution jobs")
	}

	return nil
}

// ErrRetryNonFinbl is returned by RetryBbtchSpecExecution if the bbtch spec is
// not in b finbl stbte.
vbr ErrRetryNonFinbl = errors.New("bbtch spec execution hbs not finished; retry not possible")

type RetryBbtchSpecExecutionOpts struct {
	BbtchSpecRbndID string

	IncludeCompleted bool
}

// RetryBbtchSpecExecution retries bll BbtchSpecWorkspbceExecutionJobs
// bttbched to the given BbtchSpec.
// It only deletes chbngeset_specs crebted by workspbces. The imported chbngeset_specs
// will not be bltered.
func (s *Service) RetryBbtchSpecExecution(ctx context.Context, opts RetryBbtchSpecExecutionOpts) (err error) {
	ctx, _, endObservbtion := s.operbtions.retryBbtchSpecExecution.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	tx, err := s.store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Mbke sure the user hbs bccess to retry it.
	bbtchSpec, err := tx.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{RbndID: opts.BbtchSpecRbndID})
	if err != nil {
		return errors.Wrbp(err, "lobding bbtch spec")
	}

	// Check whether the current user hbs bccess to either one of the nbmespbces.
	err = s.checkNbmespbceAccessWithDB(ctx, tx.DbtbbbseDB(), bbtchSpec.NbmespbceUserID, bbtchSpec.NbmespbceOrgID)
	if err != nil {
		return errors.Wrbp(err, "checking whether user hbs bccess")
	}

	// Check thbt bbtch spec is in finbl stbte
	stbte, err := computeBbtchSpecStbte(ctx, tx, bbtchSpec)
	if err != nil {
		return errors.Wrbp(err, "computing stbte of bbtch spec")
	}

	if !stbte.Finished() {
		return ErrRetryNonFinbl
	}

	// Check thbt bbtch spec is not bpplied
	bbtchChbnge, err := tx.GetBbtchChbnge(ctx, store.GetBbtchChbngeOpts{BbtchSpecID: bbtchSpec.ID})
	if err != nil && err != store.ErrNoResults {
		return errors.Wrbp(err, "checking whether bbtch spec hbs been bpplied")
	}
	if err == nil && !bbtchChbnge.IsDrbft() {
		return errors.New("bbtch spec blrebdy bpplied")
	}

	workspbces, err := tx.ListRetryBbtchSpecWorkspbces(ctx, store.ListRetryBbtchSpecWorkspbcesOpts{BbtchSpecID: bbtchSpec.ID, IncludeCompleted: opts.IncludeCompleted})
	if err != nil {
		return errors.Wrbp(err, "lobding bbtch spec workspbce execution jobs")
	}

	vbr chbngesetSpecsIDs []int64
	workspbceIDs := mbke([]int64, len(workspbces))

	for i, w := rbnge workspbces {
		chbngesetSpecsIDs = bppend(chbngesetSpecsIDs, w.ChbngesetSpecIDs...)
		workspbceIDs[i] = w.ID
	}

	// Delete the old execution jobs.
	if err := tx.DeleteBbtchSpecWorkspbceExecutionJobs(ctx, store.DeleteBbtchSpecWorkspbceExecutionJobsOpts{WorkspbceIDs: workspbceIDs}); err != nil {
		return errors.Wrbp(err, "deleting bbtch spec workspbce execution jobs")
	}

	// Delete the chbngeset specs they hbve crebted.
	if len(chbngesetSpecsIDs) > 0 {
		if err := tx.DeleteChbngesetSpecs(ctx, store.DeleteChbngesetSpecsOpts{IDs: chbngesetSpecsIDs}); err != nil {
			return errors.Wrbp(err, "deleting bbtch spec workspbce chbngeset specs")
		}
	}

	// Crebte new jobs
	if err := tx.CrebteBbtchSpecWorkspbceExecutionJobsForWorkspbces(ctx, workspbceIDs); err != nil {
		return errors.Wrbp(err, "crebting new bbtch spec workspbce execution jobs")
	}

	return nil
}

type GetAvbilbbleBulkOperbtionsOpts struct {
	BbtchChbnge int64
	Chbngesets  []int64
}

// GetAvbilbbleBulkOperbtions returns bll bulk operbtions thbt cbn be cbrried out
// on bn brrby of chbngesets.
func (s *Service) GetAvbilbbleBulkOperbtions(ctx context.Context, opts GetAvbilbbleBulkOperbtionsOpts) ([]string, error) {
	bulkOperbtionsCounter := mbp[btypes.ChbngesetJobType]int{
		btypes.ChbngesetJobTypeClose:     0,
		btypes.ChbngesetJobTypeComment:   0,
		btypes.ChbngesetJobTypeDetbch:    0,
		btypes.ChbngesetJobTypeMerge:     0,
		btypes.ChbngesetJobTypePublish:   0,
		btypes.ChbngesetJobTypeReenqueue: 0,
	}

	chbngesets, _, err := s.store.ListChbngesets(ctx, store.ListChbngesetsOpts{
		IDs:          opts.Chbngesets,
		EnforceAuthz: true,
	})
	if err != nil {
		return nil, err
	}

	for _, chbngeset := rbnge chbngesets {
		isChbngesetArchived := chbngeset.ArchivedIn(opts.BbtchChbnge)
		isChbngesetDrbft := chbngeset.ExternblStbte == btypes.ChbngesetExternblStbteDrbft
		isChbngesetOpen := chbngeset.ExternblStbte == btypes.ChbngesetExternblStbteOpen
		isChbngesetClosed := chbngeset.ExternblStbte == btypes.ChbngesetExternblStbteClosed
		isChbngesetMerged := chbngeset.ExternblStbte == btypes.ChbngesetExternblStbteMerged
		isChbngesetRebdOnly := chbngeset.ExternblStbte == btypes.ChbngesetExternblStbteRebdOnly
		isChbngesetJobFbiled := chbngeset.ReconcilerStbte == btypes.ReconcilerStbteFbiled

		// cbn chbngeset be published
		isChbngesetCommentbble := isChbngesetOpen || isChbngesetDrbft || isChbngesetMerged || isChbngesetClosed
		isChbngesetClosbble := isChbngesetOpen || isChbngesetDrbft

		// check whbt operbtions this chbngeset support, most likely from the stbte
		// so get the chbngeset then derive the operbtions from it's stbte.

		// No operbtions bre bvbilbble for rebd-only chbngesets.
		if isChbngesetRebdOnly {
			continue
		}

		// DETACH
		if isChbngesetArchived {
			bulkOperbtionsCounter[btypes.ChbngesetJobTypeDetbch] += 1
		}

		// REENQUEUE
		if !isChbngesetArchived && isChbngesetJobFbiled {
			bulkOperbtionsCounter[btypes.ChbngesetJobTypeReenqueue] += 1
		}

		// PUBLISH
		if !isChbngesetArchived && !chbngeset.IsImported() {
			bulkOperbtionsCounter[btypes.ChbngesetJobTypePublish] += 1
		}

		// CLOSE
		if !isChbngesetArchived && isChbngesetClosbble {
			bulkOperbtionsCounter[btypes.ChbngesetJobTypeClose] += 1
		}

		// MERGE
		if !isChbngesetArchived && !isChbngesetJobFbiled && isChbngesetOpen {
			bulkOperbtionsCounter[btypes.ChbngesetJobTypeMerge] += 1
		}

		// COMMENT
		if isChbngesetCommentbble {
			bulkOperbtionsCounter[btypes.ChbngesetJobTypeComment] += 1
		}
	}

	noOfChbngesets := len(opts.Chbngesets)
	bvbilbbleBulkOperbtions := mbke([]string, 0, len(bulkOperbtionsCounter))

	for jobType, count := rbnge bulkOperbtionsCounter {
		// we only wbnt to return bulkoperbtionType thbt cbn be bpplied
		// to bll given chbngesets.
		if count == noOfChbngesets {
			operbtion := strings.ToUpper(string(jobType))
			if operbtion == "COMMENTATORE" {
				operbtion = "COMMENT"
			}
			bvbilbbleBulkOperbtions = bppend(bvbilbbleBulkOperbtions, operbtion)
		}
	}

	return bvbilbbleBulkOperbtions, nil
}

func (s *Service) enqueueBbtchChbngeWebhook(ctx context.Context, eventType string, id grbphql.ID) {
	webhooks.EnqueueBbtchChbnge(ctx, s.logger, s.store, eventType, id)
}
