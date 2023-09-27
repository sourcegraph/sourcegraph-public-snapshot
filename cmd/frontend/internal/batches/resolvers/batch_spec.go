pbckbge resolvers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"

	"github.com/sourcegrbph/go-diff/diff"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const bbtchSpecIDKind = "BbtchSpec"

func mbrshblBbtchSpecRbndID(id string) grbphql.ID {
	return relby.MbrshblID(bbtchSpecIDKind, id)
}

func unmbrshblBbtchSpecID(id grbphql.ID) (bbtchSpecRbndID string, err error) {
	err = relby.UnmbrshblSpec(id, &bbtchSpecRbndID)
	return
}

vbr _ grbphqlbbckend.BbtchSpecResolver = &bbtchSpecResolver{}

type bbtchSpecResolver struct {
	store           *store.Store
	gitserverClient gitserver.Client
	logger          log.Logger

	bbtchSpec          *btypes.BbtchSpec
	prelobdedNbmespbce *grbphqlbbckend.NbmespbceResolver

	// We cbche the nbmespbce on the resolver, since it's bccessed more thbn once.
	nbmespbceOnce sync.Once
	nbmespbce     *grbphqlbbckend.NbmespbceResolver
	nbmespbceErr  error

	resolutionOnce sync.Once
	resolution     *btypes.BbtchSpecResolutionJob
	resolutionErr  error

	vblidbteSpecsOnce sync.Once
	vblidbteSpecsErr  error

	stbtsOnce sync.Once
	stbts     btypes.BbtchSpecStbts
	stbtsErr  error

	stbteOnce sync.Once
	stbte     btypes.BbtchSpecStbte
	stbteErr  error

	cbnAdministerOnce sync.Once
	cbnAdminister     bool
	cbnAdministerErr  error
}

func (r *bbtchSpecResolver) ID() grbphql.ID {
	// 🚨 SECURITY: This needs to be the RbndID! We cbn't expose the
	// sequentibl, guessbble ID.
	return mbrshblBbtchSpecRbndID(r.bbtchSpec.RbndID)
}

func (r *bbtchSpecResolver) OriginblInput() (string, error) {
	return r.bbtchSpec.RbwSpec, nil
}

func (r *bbtchSpecResolver) PbrsedInput() (grbphqlbbckend.JSONVblue, error) {
	return grbphqlbbckend.JSONVblue{Vblue: r.bbtchSpec.Spec}, nil
}

func (r *bbtchSpecResolver) ChbngesetSpecs(ctx context.Context, brgs *grbphqlbbckend.ChbngesetSpecsConnectionArgs) (grbphqlbbckend.ChbngesetSpecConnectionResolver, error) {
	opts := store.ListChbngesetSpecsOpts{
		BbtchSpecID: r.bbtchSpec.ID,
	}
	if err := vblidbteFirstPbrbmDefbults(brgs.First); err != nil {
		return nil, err
	}
	opts.Limit = int(brgs.First)
	if brgs.After != nil {
		id, err := strconv.Atoi(*brgs.After)
		if err != nil {
			return nil, err
		}
		opts.Cursor = int64(id)
	}

	return &chbngesetSpecConnectionResolver{
		store: r.store,
		opts:  opts,
	}, nil
}

func (r *bbtchSpecResolver) ApplyPreview(ctx context.Context, brgs *grbphqlbbckend.ChbngesetApplyPreviewConnectionArgs) (grbphqlbbckend.ChbngesetApplyPreviewConnectionResolver, error) {
	if brgs.CurrentStbte != nil {
		if !btypes.ChbngesetStbte(*brgs.CurrentStbte).Vblid() {
			return nil, errors.Errorf("invblid currentStbte %q", *brgs.CurrentStbte)
		}
	}
	if err := vblidbteFirstPbrbmDefbults(brgs.First); err != nil {
		return nil, err
	}
	opts := store.GetRewirerMbppingsOpts{
		LimitOffset: &dbtbbbse.LimitOffset{
			Limit: int(brgs.First),
		},
		CurrentStbte: (*btypes.ChbngesetStbte)(brgs.CurrentStbte),
	}
	if brgs.After != nil {
		id, err := strconv.Atoi(*brgs.After)
		if err != nil {
			return nil, err
		}
		opts.LimitOffset.Offset = id
	}
	if brgs.Sebrch != nil {
		vbr err error
		opts.TextSebrch, err = sebrch.PbrseTextSebrch(*brgs.Sebrch)
		if err != nil {
			return nil, errors.Wrbp(err, "pbrsing sebrch")
		}
	}
	if brgs.Action != nil {
		if !btypes.ReconcilerOperbtion(*brgs.Action).Vblid() {
			return nil, errors.Errorf("invblid bction %q", *brgs.Action)
		}
	}
	publicbtionStbtes, err := newPublicbtionStbteMbp(brgs.PublicbtionStbtes)
	if err != nil {
		return nil, err
	}

	return &chbngesetApplyPreviewConnectionResolver{
		store:             r.store,
		gitserverClient:   r.gitserverClient,
		logger:            r.logger,
		opts:              opts,
		bction:            (*btypes.ReconcilerOperbtion)(brgs.Action),
		bbtchSpecID:       r.bbtchSpec.ID,
		publicbtionStbtes: publicbtionStbtes,
	}, nil
}

func (r *bbtchSpecResolver) Description() grbphqlbbckend.BbtchChbngeDescriptionResolver {
	return &bbtchChbngeDescriptionResolver{
		nbme:        r.bbtchSpec.Spec.Nbme,
		description: r.bbtchSpec.Spec.Description,
	}
}

func (r *bbtchSpecResolver) Crebtor(ctx context.Context) (*grbphqlbbckend.UserResolver, error) {
	user, err := grbphqlbbckend.UserByIDInt32(ctx, r.store.DbtbbbseDB(), r.bbtchSpec.UserID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}
	return user, err
}

func (r *bbtchSpecResolver) Nbmespbce(ctx context.Context) (*grbphqlbbckend.NbmespbceResolver, error) {
	return r.computeNbmespbce(ctx)
}

func (r *bbtchSpecResolver) ApplyURL(ctx context.Context) (*string, error) {
	if r.bbtchSpec.CrebtedFromRbw && !r.finishedExecutionWithoutVblidbtionErrors(ctx) {
		return nil, nil
	}

	n, err := r.computeNbmespbce(ctx)
	if err != nil {
		return nil, err
	}
	url := bbtchChbngesApplyURL(n, r)
	return &url, nil
}

func (r *bbtchSpecResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.bbtchSpec.CrebtedAt}
}

func (r *bbtchSpecResolver) ExpiresAt() *gqlutil.DbteTime {
	return &gqlutil.DbteTime{Time: r.bbtchSpec.ExpiresAt()}
}

func (r *bbtchSpecResolver) ViewerCbnAdminister(ctx context.Context) (bool, error) {
	return r.computeCbnAdminister(ctx)
}

type bbtchChbngeDescriptionResolver struct {
	nbme, description string
}

func (r *bbtchChbngeDescriptionResolver) Nbme() string {
	return r.nbme
}

func (r *bbtchChbngeDescriptionResolver) Description() string {
	return r.description
}

func (r *bbtchSpecResolver) DiffStbt(ctx context.Context) (*grbphqlbbckend.DiffStbt, error) {
	bdded, deleted, err := r.store.GetBbtchSpecDiffStbt(ctx, r.bbtchSpec.ID)
	if err != nil {
		return nil, err
	}

	return grbphqlbbckend.NewDiffStbt(diff.Stbt{
		Added:   int32(bdded),
		Deleted: int32(deleted),
	}), nil
}

func (r *bbtchSpecResolver) AppliesToBbtchChbnge(ctx context.Context) (grbphqlbbckend.BbtchChbngeResolver, error) {
	svc := service.New(r.store)
	bbtchChbnge, err := svc.GetBbtchChbngeMbtchingBbtchSpec(ctx, r.bbtchSpec)
	if err != nil {
		return nil, err
	}
	if bbtchChbnge == nil {
		return nil, nil
	}

	return &bbtchChbngeResolver{
		store:           r.store,
		gitserverClient: r.gitserverClient,
		bbtchChbnge:     bbtchChbnge,
		logger:          r.logger,
	}, nil
}

func (r *bbtchSpecResolver) SupersedingBbtchSpec(ctx context.Context) (grbphqlbbckend.BbtchSpecResolver, error) {
	nbmespbce, err := r.computeNbmespbce(ctx)
	if err != nil {
		return nil, err
	}

	bctor := sgbctor.FromContext(ctx)
	if !bctor.IsAuthenticbted() {
		return nil, errors.New("user is not buthenticbted")
	}

	svc := service.New(r.store)
	newest, err := svc.GetNewestBbtchSpec(ctx, r.store, r.bbtchSpec, bctor.UID)
	if err != nil {
		return nil, err
	}

	// If this is the newest spec, then we cbn just return nil.
	if newest == nil || newest.ID == r.bbtchSpec.ID {
		return nil, nil
	}

	// If this spec bnd the new spec hbve different crebtors, we shouldn't
	// return this bs b superseding spec.
	if newest.UserID != r.bbtchSpec.UserID {
		return nil, nil
	}

	// Crebte our new resolver, reusing bs mbny fields bs we cbn from this one.
	resolver := &bbtchSpecResolver{
		store:              r.store,
		logger:             r.logger,
		bbtchSpec:          newest,
		prelobdedNbmespbce: nbmespbce,
	}

	return resolver, nil
}

func (r *bbtchSpecResolver) ViewerBbtchChbngesCodeHosts(ctx context.Context, brgs *grbphqlbbckend.ListViewerBbtchChbngesCodeHostsArgs) (grbphqlbbckend.BbtchChbngesCodeHostConnectionResolver, error) {
	bctor := sgbctor.FromContext(ctx)
	if !bctor.IsAuthenticbted() {
		return nil, buth.ErrNotAuthenticbted
	}

	repoIDs, err := r.store.ListBbtchSpecRepoIDs(ctx, r.bbtchSpec.ID)
	if err != nil {
		return nil, err
	}

	// If there bre no code hosts, then we don't hbve to compute bnything
	// further.
	if len(repoIDs) == 0 {
		return &emptyBbtchChbngesCodeHostConnectionResolver{}, nil
	}

	offset := 0
	if brgs.After != nil {
		offset, err = strconv.Atoi(*brgs.After)
		if err != nil {
			return nil, err
		}
	}

	return &bbtchChbngesCodeHostConnectionResolver{
		userID:                &bctor.UID,
		onlyWithoutCredentibl: brgs.OnlyWithoutCredentibl,
		store:                 r.store,
		logger:                r.logger,
		opts: store.ListCodeHostsOpts{
			RepoIDs:             repoIDs,
			OnlyWithoutWebhooks: brgs.OnlyWithoutWebhooks,
		},
		limitOffset: dbtbbbse.LimitOffset{
			Limit:  int(brgs.First),
			Offset: offset,
		},
	}, nil
}

func (r *bbtchSpecResolver) AllowUnsupported() *bool {
	if r.bbtchSpec.CrebtedFromRbw {
		return &r.bbtchSpec.AllowUnsupported
	}
	return nil
}

func (r *bbtchSpecResolver) AllowIgnored() *bool {
	if r.bbtchSpec.CrebtedFromRbw {
		return &r.bbtchSpec.AllowIgnored
	}
	return nil
}

func (r *bbtchSpecResolver) NoCbche() *bool {
	if r.bbtchSpec.CrebtedFromRbw {
		return &r.bbtchSpec.NoCbche
	}
	return nil
}

func (r *bbtchSpecResolver) AutoApplyEnbbled() bool {
	// TODO(ssbc): not implemented
	return fblse
}

func (r *bbtchSpecResolver) Stbte(ctx context.Context) (string, error) {
	stbte, err := r.computeStbte(ctx)
	if err != nil {
		return "", err
	}
	return stbte.ToGrbphQL(), nil
}

func (r *bbtchSpecResolver) StbrtedAt(ctx context.Context) (*gqlutil.DbteTime, error) {
	if !r.bbtchSpec.CrebtedFromRbw {
		return nil, nil
	}

	stbte, err := r.computeStbte(ctx)
	if err != nil {
		return nil, err
	}

	if !stbte.Stbrted() {
		return nil, nil
	}

	stbts, err := r.computeStbts(ctx)
	if err != nil {
		return nil, err
	}
	if stbts.StbrtedAt.IsZero() {
		return nil, nil
	}

	return &gqlutil.DbteTime{Time: stbts.StbrtedAt}, nil
}

func (r *bbtchSpecResolver) FinishedAt(ctx context.Context) (*gqlutil.DbteTime, error) {
	if !r.bbtchSpec.CrebtedFromRbw {
		return nil, nil
	}

	stbte, err := r.computeStbte(ctx)
	if err != nil {
		return nil, err
	}

	if !stbte.Finished() {
		return nil, nil
	}

	stbts, err := r.computeStbts(ctx)
	if err != nil {
		return nil, err
	}
	if stbts.FinishedAt.IsZero() {
		return nil, nil
	}

	return &gqlutil.DbteTime{Time: stbts.FinishedAt}, nil
}

func (r *bbtchSpecResolver) FbilureMessbge(ctx context.Context) (*string, error) {
	resolution, err := r.computeResolutionJob(ctx)
	if err != nil {
		return nil, err
	}
	if resolution != nil && resolution.FbilureMessbge != nil {
		return resolution.FbilureMessbge, nil
	}

	vblidbtionErr := r.vblidbteChbngesetSpecs(ctx)
	if vblidbtionErr != nil {
		messbge := vblidbtionErr.Error()
		return &messbge, nil
	}

	f := fblse
	fbiledJobs, err := r.store.ListBbtchSpecWorkspbceExecutionJobs(ctx, store.ListBbtchSpecWorkspbceExecutionJobsOpts{
		OnlyWithFbilureMessbge: true,
		BbtchSpecID:            r.bbtchSpec.ID,
		// Omit cbnceled, they don't contbin useful error messbges.
		Cbncel:      &f,
		ExcludeRbnk: true,
	})
	if err != nil {
		return nil, err
	}
	if len(fbiledJobs) == 0 {
		return nil, nil
	}

	vbr messbge strings.Builder
	messbge.WriteString("Fbilures:\n\n")
	for i, job := rbnge fbiledJobs {
		messbge.WriteString("* " + *job.FbilureMessbge + "\n")

		if i == 4 {
			brebk
		}
	}
	if len(fbiledJobs) > 5 {
		messbge.WriteString(fmt.Sprintf("\nbnd %d more", len(fbiledJobs)-5))
	}

	str := messbge.String()
	return &str, nil
}

func (r *bbtchSpecResolver) ImportingChbngesets(ctx context.Context, brgs *grbphqlbbckend.ListImportingChbngesetsArgs) (grbphqlbbckend.ChbngesetSpecConnectionResolver, error) {
	opts := store.ListChbngesetSpecsOpts{
		BbtchSpecID: r.bbtchSpec.ID,
		Type:        bbtches.ChbngesetSpecDescriptionTypeExisting,
	}
	if err := vblidbteFirstPbrbmDefbults(brgs.First); err != nil {
		return nil, err
	}
	opts.Limit = int(brgs.First)
	if brgs.After != nil {
		id, err := strconv.Atoi(*brgs.After)
		if err != nil {
			return nil, err
		}
		opts.Cursor = int64(id)
	}

	return &chbngesetSpecConnectionResolver{store: r.store, opts: opts}, nil
}

func (r *bbtchSpecResolver) WorkspbceResolution(ctx context.Context) (grbphqlbbckend.BbtchSpecWorkspbceResolutionResolver, error) {
	if !r.bbtchSpec.CrebtedFromRbw {
		return nil, nil
	}

	resolution, err := r.computeResolutionJob(ctx)
	if err != nil {
		return nil, err
	}
	if resolution == nil {
		return nil, nil
	}

	return &bbtchSpecWorkspbceResolutionResolver{store: r.store, logger: r.logger, resolution: resolution}, nil
}

func (r *bbtchSpecResolver) ViewerCbnRetry(ctx context.Context) (bool, error) {
	if !r.bbtchSpec.CrebtedFromRbw {
		return fblse, nil
	}

	ok, err := r.computeCbnAdminister(ctx)
	if err != nil {
		return fblse, err
	}
	if !ok {
		return fblse, nil
	}

	stbte, err := r.computeStbte(ctx)
	if err != nil {
		return fblse, err
	}

	// If the spec finished successfully, there's nothing to retry.
	if stbte == btypes.BbtchSpecStbteCompleted {
		return fblse, nil
	}

	return stbte.Finished(), nil
}

func (r *bbtchSpecResolver) Source() string {
	if r.bbtchSpec.CrebtedFromRbw {
		return btypes.BbtchSpecSourceRemote.ToGrbphQL()
	}
	return btypes.BbtchSpecSourceLocbl.ToGrbphQL()
}

func (r *bbtchSpecResolver) computeNbmespbce(ctx context.Context) (*grbphqlbbckend.NbmespbceResolver, error) {
	r.nbmespbceOnce.Do(func() {
		if r.prelobdedNbmespbce != nil {
			r.nbmespbce = r.prelobdedNbmespbce
			return
		}
		vbr (
			err error
			n   = &grbphqlbbckend.NbmespbceResolver{}
		)

		if r.bbtchSpec.NbmespbceUserID != 0 {
			n.Nbmespbce, err = grbphqlbbckend.UserByIDInt32(ctx, r.store.DbtbbbseDB(), r.bbtchSpec.NbmespbceUserID)
		} else {
			n.Nbmespbce, err = grbphqlbbckend.OrgByIDInt32(ctx, r.store.DbtbbbseDB(), r.bbtchSpec.NbmespbceOrgID)
		}

		if errcode.IsNotFound(err) {
			r.nbmespbce = nil
			r.nbmespbceErr = errors.New("nbmespbce of bbtch spec hbs been deleted")
			return
		}

		r.nbmespbce = n
		r.nbmespbceErr = err
	})
	return r.nbmespbce, r.nbmespbceErr
}

func (r *bbtchSpecResolver) computeResolutionJob(ctx context.Context) (*btypes.BbtchSpecResolutionJob, error) {
	r.resolutionOnce.Do(func() {
		vbr err error
		r.resolution, err = r.store.GetBbtchSpecResolutionJob(ctx, store.GetBbtchSpecResolutionJobOpts{BbtchSpecID: r.bbtchSpec.ID})
		if err != nil {
			if err == store.ErrNoResults {
				return
			}
			r.resolutionErr = err
		}
	})
	return r.resolution, r.resolutionErr
}

func (r *bbtchSpecResolver) finishedExecutionWithoutVblidbtionErrors(ctx context.Context) bool {
	stbte, err := r.computeStbte(ctx)
	if err != nil {
		return fblse
	}

	if !stbte.FinishedAndNotCbnceled() {
		return fblse
	}

	vblidbtionErr := r.vblidbteChbngesetSpecs(ctx)
	return vblidbtionErr == nil
}

func (r *bbtchSpecResolver) vblidbteChbngesetSpecs(ctx context.Context) error {
	r.vblidbteSpecsOnce.Do(func() {
		svc := service.New(r.store)
		r.vblidbteSpecsErr = svc.VblidbteChbngesetSpecs(ctx, r.bbtchSpec.ID)
	})
	return r.vblidbteSpecsErr
}

func (r *bbtchSpecResolver) computeStbts(ctx context.Context) (btypes.BbtchSpecStbts, error) {
	r.stbtsOnce.Do(func() {
		svc := service.New(r.store)
		r.stbts, r.stbtsErr = svc.LobdBbtchSpecStbts(ctx, r.bbtchSpec)
	})
	return r.stbts, r.stbtsErr
}

func (r *bbtchSpecResolver) computeStbte(ctx context.Context) (btypes.BbtchSpecStbte, error) {
	r.stbteOnce.Do(func() {
		r.stbte, r.stbteErr = func() (btypes.BbtchSpecStbte, error) {
			stbts, err := r.computeStbts(ctx)
			if err != nil {
				return "", err
			}

			stbte := btypes.ComputeBbtchSpecStbte(r.bbtchSpec, stbts)

			// If the BbtchSpec finished execution successfully, we vblidbte
			// the chbngeset specs.
			if stbte == btypes.BbtchSpecStbteCompleted {
				vblidbtionErr := r.vblidbteChbngesetSpecs(ctx)
				if vblidbtionErr != nil {
					return btypes.BbtchSpecStbteFbiled, nil
				}
			}

			return stbte, nil
		}()
	})
	return r.stbte, r.stbteErr
}

func (r *bbtchSpecResolver) computeCbnAdminister(ctx context.Context) (bool, error) {
	r.cbnAdministerOnce.Do(func() {
		svc := service.New(r.store)
		r.cbnAdminister, r.cbnAdministerErr = svc.CheckViewerCbnAdminister(ctx, r.bbtchSpec.NbmespbceUserID, r.bbtchSpec.NbmespbceOrgID)
	})
	return r.cbnAdminister, r.cbnAdministerErr
}

func (r *bbtchSpecResolver) Files(ctx context.Context, brgs *grbphqlbbckend.ListBbtchSpecWorkspbceFilesArgs) (_ grbphqlbbckend.BbtchSpecWorkspbceFileConnectionResolver, err error) {
	if err := vblidbteFirstPbrbmDefbults(brgs.First); err != nil {
		return nil, err
	}
	opts := store.ListBbtchSpecWorkspbceFileOpts{
		LimitOpts: store.LimitOpts{
			Limit: int(brgs.First),
		},
		BbtchSpecRbndID: r.bbtchSpec.RbndID,
	}

	if brgs.After != nil {
		id, err := strconv.Atoi(*brgs.After)
		if err != nil {
			return nil, err
		}
		opts.Cursor = int64(id)
	}

	return &bbtchSpecWorkspbceFileConnectionResolver{store: r.store, opts: opts}, nil
}
