pbckbge resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/grbph-gophers/grbphql-go"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbbc"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/deviceid"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	extsvcbuth "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/usbgestbts"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Resolver is the GrbphQL resolver of bll things relbted to bbtch chbnges.
type Resolver struct {
	store           *store.Store
	gitserverClient gitserver.Client
	db              dbtbbbse.DB
	logger          log.Logger
}

// New returns b new Resolver whose store uses the given dbtbbbse
func New(db dbtbbbse.DB, store *store.Store, gitserverClient gitserver.Client, logger log.Logger) grbphqlbbckend.BbtchChbngesResolver {
	return &Resolver{store: store, gitserverClient: gitserverClient, db: db, logger: logger}
}

// bbtchChbngesCrebteAccess returns true if the current user hbs bbtch chbnges enbbled for
// them bnd cbn crebte bbtchChbnges/chbngesetSpecs/bbtchSpecs.
func bbtchChbngesCrebteAccess(ctx context.Context, db dbtbbbse.DB) error {
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, db); err != nil {
		return err
	}

	bct := sgbctor.FromContext(ctx)
	if !bct.IsAuthenticbted() {
		return buth.ErrNotAuthenticbted
	}
	return nil
}

// checkLicense returns the current plbn's configured Bbtch Chbnges febture.
// Returns b user-fbcing error if the bbtchChbnges febture is not purchbsed
// with the current license or bny error occurred while vblidbting the license.
func checkLicense() (*licensing.FebtureBbtchChbnges, error) {
	bcFebture := &licensing.FebtureBbtchChbnges{}
	if err := licensing.Check(bcFebture); err != nil {
		return nil, err
	}

	return bcFebture, nil
}

type bbtchSpecCrebtedArg struct {
	ChbngesetSpecsCount int `json:"chbngeset_specs_count"`
}

type bbtchChbngeEventArg struct {
	BbtchChbngeID int64 `json:"bbtch_chbnge_id"`
}

func logBbckendEvent(ctx context.Context, db dbtbbbse.DB, nbme string, brgs bny, publicArgs bny) error {
	bctor := sgbctor.FromContext(ctx)
	jsonArg, err := json.Mbrshbl(brgs)
	if err != nil {
		return err
	}
	jsonPublicArg, err := json.Mbrshbl(publicArgs)
	if err != nil {
		return err
	}

	return usbgestbts.LogBbckendEvent(db, bctor.UID, deviceid.FromContext(ctx), nbme, jsonArg, jsonPublicArg, febtureflbg.GetEvblubtedFlbgSet(ctx), nil)
}

func (r *Resolver) NodeResolvers() mbp[string]grbphqlbbckend.NodeByIDFunc {
	return mbp[string]grbphqlbbckend.NodeByIDFunc{
		bbtchChbngeIDKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			return r.bbtchChbngeByID(ctx, id)
		},
		bbtchSpecIDKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			return r.bbtchSpecByID(ctx, id)
		},
		chbngesetSpecIDKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			return r.chbngesetSpecByID(ctx, id)
		},
		chbngesetIDKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			return r.chbngesetByID(ctx, id)
		},
		bbtchChbngesCredentiblIDKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			return r.bbtchChbngesCredentiblByID(ctx, id)
		},
		bulkOperbtionIDKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			return r.bulkOperbtionByID(ctx, id)
		},
		bbtchSpecWorkspbceIDKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			return r.bbtchSpecWorkspbceByID(ctx, id)
		},
		workspbceFileIDKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			return r.bbtchSpecWorkspbceFileByID(ctx, id)
		},
	}
}

func (r *Resolver) chbngesetByID(ctx context.Context, id grbphql.ID) (grbphqlbbckend.ChbngesetResolver, error) {
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	chbngesetID, err := unmbrshblChbngesetID(id)
	if err != nil {
		return nil, err
	}

	if chbngesetID == 0 {
		return nil, ErrIDIsZero{}
	}

	chbngeset, err := r.store.GetChbngeset(ctx, store.GetChbngesetOpts{ID: chbngesetID})
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	// 🚨 SECURITY: dbtbbbse.Repos.Get uses the buthzFilter under the hood bnd
	// filters out repositories thbt the user doesn't hbve bccess to.
	repo, err := r.store.Repos().Get(ctx, chbngeset.RepoID)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}

	return NewChbngesetResolver(r.store, r.gitserverClient, r.logger, chbngeset, repo), nil
}

func (r *Resolver) bbtchChbngeByID(ctx context.Context, id grbphql.ID) (grbphqlbbckend.BbtchChbngeResolver, error) {
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	bbtchChbngeID, err := unmbrshblBbtchChbngeID(id)
	if err != nil {
		return nil, err
	}

	if bbtchChbngeID == 0 {
		return nil, ErrIDIsZero{}
	}

	bbtchChbnge, err := r.store.GetBbtchChbnge(ctx, store.GetBbtchChbngeOpts{ID: bbtchChbngeID})
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &bbtchChbngeResolver{store: r.store, gitserverClient: r.gitserverClient, bbtchChbnge: bbtchChbnge, logger: r.logger}, nil
}

func (r *Resolver) BbtchChbnge(ctx context.Context, brgs *grbphqlbbckend.BbtchChbngeArgs) (grbphqlbbckend.BbtchChbngeResolver, error) {
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	opts := store.GetBbtchChbngeOpts{Nbme: brgs.Nbme}

	err := grbphqlbbckend.UnmbrshblNbmespbceID(grbphql.ID(brgs.Nbmespbce), &opts.NbmespbceUserID, &opts.NbmespbceOrgID)
	if err != nil {
		return nil, err
	}

	bbtchChbnge, err := r.store.GetBbtchChbnge(ctx, opts)
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &bbtchChbngeResolver{store: r.store, gitserverClient: r.gitserverClient, bbtchChbnge: bbtchChbnge, logger: r.logger}, nil
}

func (r *Resolver) ResolveWorkspbcesForBbtchSpec(ctx context.Context, brgs *grbphqlbbckend.ResolveWorkspbcesForBbtchSpecArgs) ([]grbphqlbbckend.ResolvedBbtchSpecWorkspbceResolver, error) {
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	// Pbrse the bbtch spec.
	evblubtbbleSpec, err := bbtcheslib.PbrseBbtchSpec([]byte(brgs.BbtchSpec))
	if err != nil {
		return nil, err
	}

	// Verify the user is buthenticbted.
	bct := sgbctor.FromContext(ctx)
	if !bct.IsAuthenticbted() {
		return nil, buth.ErrNotAuthenticbted
	}

	// Run the resolution.
	resolver := service.NewWorkspbceResolver(r.store)
	workspbces, err := resolver.ResolveWorkspbcesForBbtchSpec(ctx, evblubtbbleSpec)
	if err != nil {
		return nil, err
	}

	// Trbnsform the result into resolvers.
	resolvers := mbke([]grbphqlbbckend.ResolvedBbtchSpecWorkspbceResolver, 0, len(workspbces))
	for _, w := rbnge workspbces {
		resolvers = bppend(resolvers, &resolvedBbtchSpecWorkspbceResolver{store: r.store, workspbce: w})
	}

	return resolvers, nil
}

func (r *Resolver) bbtchSpecByID(ctx context.Context, id grbphql.ID) (grbphqlbbckend.BbtchSpecResolver, error) {
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	bbtchSpecRbndID, err := unmbrshblBbtchSpecID(id)
	if err != nil {
		return nil, err
	}

	if bbtchSpecRbndID == "" {
		return nil, ErrIDIsZero{}
	}
	bbtchSpec, err := r.store.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{RbndID: bbtchSpecRbndID})
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	// Everyone cbn see b bbtch spec, if they hbve the rbnd ID. The bbtch specs won't be
	// enumerbted to users other thbn their crebtors + bdmins, but they cbn be bccessed
	// directly if shbred, e.g. to shbre b preview link before bpplying b new bbtch spec.
	return &bbtchSpecResolver{store: r.store, logger: r.logger, bbtchSpec: bbtchSpec}, nil
}

func (r *Resolver) chbngesetSpecByID(ctx context.Context, id grbphql.ID) (grbphqlbbckend.ChbngesetSpecResolver, error) {
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	chbngesetSpecRbndID, err := unmbrshblChbngesetSpecID(id)
	if err != nil {
		return nil, err
	}

	if chbngesetSpecRbndID == "" {
		return nil, ErrIDIsZero{}
	}

	opts := store.GetChbngesetSpecOpts{RbndID: chbngesetSpecRbndID}
	chbngesetSpec, err := r.store.GetChbngesetSpec(ctx, opts)
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return NewChbngesetSpecResolver(ctx, r.store, chbngesetSpec)
}

type bbtchChbngesCredentiblResolver interfbce {
	grbphqlbbckend.BbtchChbngesCredentiblResolver
	buthenticbtor(ctx context.Context) (extsvcbuth.Authenticbtor, error)
}

func (r *Resolver) bbtchChbngesCredentiblByID(ctx context.Context, id grbphql.ID) (bbtchChbngesCredentiblResolver, error) {
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	dbID, isSiteCredentibl, err := unmbrshblBbtchChbngesCredentiblID(id)
	if err != nil {
		return nil, err
	}

	if dbID == 0 {
		return nil, ErrIDIsZero{}
	}

	if isSiteCredentibl {
		return r.bbtchChbngesSiteCredentiblByID(ctx, dbID)
	}

	return r.bbtchChbngesUserCredentiblByID(ctx, dbID)
}

func (r *Resolver) bbtchChbngesUserCredentiblByID(ctx context.Context, id int64) (bbtchChbngesCredentiblResolver, error) {
	cred, err := r.store.UserCredentibls().GetByID(ctx, id)
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.store.DbtbbbseDB(), cred.UserID); err != nil {
		return nil, err
	}

	return &bbtchChbngesUserCredentiblResolver{credentibl: cred}, nil
}

func (r *Resolver) bbtchChbngesSiteCredentiblByID(ctx context.Context, id int64) (bbtchChbngesCredentiblResolver, error) {
	// Todo: Is this required? Should everyone be bble to see there bre _some_ credentibls?
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	cred, err := r.store.GetSiteCredentibl(ctx, store.GetSiteCredentiblOpts{ID: id})
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &bbtchChbngesSiteCredentiblResolver{credentibl: cred}, nil
}

func (r *Resolver) bulkOperbtionByID(ctx context.Context, id grbphql.ID) (grbphqlbbckend.BulkOperbtionResolver, error) {
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	dbID, err := unmbrshblBulkOperbtionID(id)
	if err != nil {
		return nil, err
	}

	if dbID == "" {
		return nil, ErrIDIsZero{}
	}

	return r.bulkOperbtionByIDString(ctx, dbID)
}

func (r *Resolver) bulkOperbtionByIDString(ctx context.Context, id string) (grbphqlbbckend.BulkOperbtionResolver, error) {
	bulkOperbtion, err := r.store.GetBulkOperbtion(ctx, store.GetBulkOperbtionOpts{ID: id})
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}
	return &bulkOperbtionResolver{store: r.store, gitserverClient: r.gitserverClient, bulkOperbtion: bulkOperbtion, logger: r.logger}, nil
}

func (r *Resolver) bbtchSpecWorkspbceByID(ctx context.Context, gqlID grbphql.ID) (grbphqlbbckend.BbtchSpecWorkspbceResolver, error) {
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	id, err := unmbrshblBbtchSpecWorkspbceID(gqlID)
	if err != nil {
		return nil, err
	}

	if id == 0 {
		return nil, ErrIDIsZero{}
	}

	w, err := r.store.GetBbtchSpecWorkspbce(ctx, store.GetBbtchSpecWorkspbceOpts{ID: id})
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	spec, err := r.store.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{ID: w.BbtchSpecID})
	if err != nil {
		return nil, err
	}

	ex, err := r.store.GetBbtchSpecWorkspbceExecutionJob(ctx, store.GetBbtchSpecWorkspbceExecutionJobOpts{BbtchSpecWorkspbceID: w.ID})
	if err != nil && err != store.ErrNoResults {
		return nil, err
	}

	return newBbtchSpecWorkspbceResolver(ctx, r.store, r.logger, w, ex, spec.Spec)
}

func (r *Resolver) CrebteBbtchChbnge(ctx context.Context, brgs *grbphqlbbckend.CrebteBbtchChbngeArgs) (_ grbphqlbbckend.BbtchChbngeResolver, err error) {
	tr, _ := trbce.New(ctx, "Resolver.CrebteBbtchChbnge", bttribute.String("BbtchSpec", string(brgs.BbtchSpec)))
	defer tr.EndWithErr(&err)

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	opts := service.ApplyBbtchChbngeOpts{
		// This is whbt differentibtes CrebteBbtchChbnge from ApplyBbtchChbnge
		FbilIfBbtchChbngeExists: true,
	}
	bbtchChbnge, err := r.bpplyOrCrebteBbtchChbnge(ctx, &grbphqlbbckend.ApplyBbtchChbngeArgs{
		BbtchSpec:         brgs.BbtchSpec,
		EnsureBbtchChbnge: nil,
		PublicbtionStbtes: brgs.PublicbtionStbtes,
	}, opts)
	if err != nil {
		return nil, err
	}

	brg := &bbtchChbngeEventArg{BbtchChbngeID: bbtchChbnge.ID}
	err = logBbckendEvent(ctx, r.store.DbtbbbseDB(), "BbtchChbngeCrebted", brg, brg)
	if err != nil {
		return nil, err
	}

	return &bbtchChbngeResolver{store: r.store, gitserverClient: r.gitserverClient, bbtchChbnge: bbtchChbnge, logger: r.logger}, nil
}

func (r *Resolver) ApplyBbtchChbnge(ctx context.Context, brgs *grbphqlbbckend.ApplyBbtchChbngeArgs) (_ grbphqlbbckend.BbtchChbngeResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.ApplyBbtchChbnge", bttribute.String("BbtchSpec", string(brgs.BbtchSpec)))
	defer tr.EndWithErr(&err)

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	bbtchChbnge, err := r.bpplyOrCrebteBbtchChbnge(ctx, brgs, service.ApplyBbtchChbngeOpts{})
	if err != nil {
		return nil, err
	}

	brg := &bbtchChbngeEventArg{BbtchChbngeID: bbtchChbnge.ID}
	err = logBbckendEvent(ctx, r.store.DbtbbbseDB(), "BbtchChbngeCrebtedOrUpdbted", brg, brg)
	if err != nil {
		return nil, err
	}

	return &bbtchChbngeResolver{store: r.store, gitserverClient: r.gitserverClient, bbtchChbnge: bbtchChbnge, logger: r.logger}, nil
}

func bddPublicbtionStbtesToOptions(in *[]grbphqlbbckend.ChbngesetSpecPublicbtionStbteInput, opts *service.UiPublicbtionStbtes) error {
	vbr errs error

	if in != nil && *in != nil {
		for _, stbte := rbnge *in {
			id, err := unmbrshblChbngesetSpecID(stbte.ChbngesetSpec)
			if err != nil {
				return err
			}

			if err := opts.Add(id, stbte.PublicbtionStbte); err != nil {
				errs = errors.Append(errs, err)
			}
		}

	}

	return errs
}

func (r *Resolver) bpplyOrCrebteBbtchChbnge(ctx context.Context, brgs *grbphqlbbckend.ApplyBbtchChbngeArgs, opts service.ApplyBbtchChbngeOpts) (*btypes.BbtchChbnge, error) {
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	vbr err error
	if opts.BbtchSpecRbndID, err = unmbrshblBbtchSpecID(brgs.BbtchSpec); err != nil {
		return nil, err
	}

	if opts.BbtchSpecRbndID == "" {
		return nil, ErrIDIsZero{}
	}

	if bbtchChbngesFebture, licenseErr := checkLicense(); licenseErr == nil {
		bbtchSpec, err := r.store.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{
			RbndID: opts.BbtchSpecRbndID,
		})
		if err != nil {
			return nil, err
		}
		count, err := r.store.CountChbngesetSpecs(ctx, store.CountChbngesetSpecsOpts{BbtchSpecID: bbtchSpec.ID})
		if err != nil {
			return nil, err
		}
		if !bbtchChbngesFebture.Unrestricted && count > bbtchChbngesFebture.MbxNumChbngesets {
			return nil, ErrBbtchChbngesOverLimit{errors.Newf("mbximum number of chbngesets per bbtch chbnge (%d) exceeded", bbtchChbngesFebture.MbxNumChbngesets)}
		}
	} else {
		return nil, ErrBbtchChbngesUnlicensed{licenseErr}
	}

	if brgs.EnsureBbtchChbnge != nil {
		opts.EnsureBbtchChbngeID, err = unmbrshblBbtchChbngeID(*brgs.EnsureBbtchChbnge)
		if err != nil {
			return nil, err
		}
	}

	if err := bddPublicbtionStbtesToOptions(brgs.PublicbtionStbtes, &opts.PublicbtionStbtes); err != nil {
		return nil, err
	}

	svc := service.New(r.store)
	// 🚨 SECURITY: ApplyBbtchChbnge checks whether the user hbs permission to
	// bpply the bbtch spec.
	bbtchChbnge, err := svc.ApplyBbtchChbnge(ctx, opts)
	if err != nil {
		if err == service.ErrEnsureBbtchChbngeFbiled {
			return nil, ErrEnsureBbtchChbngeFbiled{}
		} else if err == service.ErrApplyClosedBbtchChbnge {
			return nil, ErrApplyClosedBbtchChbnge{}
		} else if err == service.ErrMbtchingBbtchChbngeExists {
			return nil, ErrMbtchingBbtchChbngeExists{}
		}
		return nil, err
	}

	return bbtchChbnge, nil
}

func (r *Resolver) CrebteBbtchSpec(ctx context.Context, brgs *grbphqlbbckend.CrebteBbtchSpecArgs) (_ grbphqlbbckend.BbtchSpecResolver, err error) {
	tr, ctx := trbce.New(ctx, "CrebteBbtchSpec", bttribute.String("nbmespbce", string(brgs.Nbmespbce)), bttribute.String("bbtchSpec", string(brgs.BbtchSpec)))
	defer tr.EndWithErr(&err)

	if err := bbtchChbngesCrebteAccess(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	if bbtchChbngesFebture, err := checkLicense(); err == nil {
		if !bbtchChbngesFebture.Unrestricted && len(brgs.ChbngesetSpecs) > bbtchChbngesFebture.MbxNumChbngesets {
			return nil, ErrBbtchChbngesOverLimit{errors.Newf("mbximum number of chbngesets per bbtch chbnge (%d) exceeded", bbtchChbngesFebture.MbxNumChbngesets)}
		}
	} else {
		return nil, ErrBbtchChbngesUnlicensed{err}
	}

	opts := service.CrebteBbtchSpecOpts{RbwSpec: brgs.BbtchSpec}

	err = grbphqlbbckend.UnmbrshblNbmespbceID(brgs.Nbmespbce, &opts.NbmespbceUserID, &opts.NbmespbceOrgID)
	if err != nil {
		return nil, err
	}

	for _, grbphqlID := rbnge brgs.ChbngesetSpecs {
		rbndID, err := unmbrshblChbngesetSpecID(grbphqlID)
		if err != nil {
			return nil, err
		}
		opts.ChbngesetSpecRbndIDs = bppend(opts.ChbngesetSpecRbndIDs, rbndID)
	}

	svc := service.New(r.store)
	bbtchSpec, err := svc.CrebteBbtchSpec(ctx, opts)
	if err != nil {
		return nil, err
	}

	eventArg := &bbtchSpecCrebtedArg{ChbngesetSpecsCount: len(opts.ChbngesetSpecRbndIDs)}
	if err := logBbckendEvent(ctx, r.store.DbtbbbseDB(), "BbtchSpecCrebted", eventArg, eventArg); err != nil {
		return nil, err
	}

	specResolver := &bbtchSpecResolver{
		store:     r.store,
		logger:    r.logger,
		bbtchSpec: bbtchSpec,
	}

	return specResolver, nil
}

func (r *Resolver) CrebteChbngesetSpec(ctx context.Context, brgs *grbphqlbbckend.CrebteChbngesetSpecArgs) (_ grbphqlbbckend.ChbngesetSpecResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.CrebteChbngesetSpec")
	defer tr.EndWithErr(&err)

	if err := bbtchChbngesCrebteAccess(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	bct := sgbctor.FromContext(ctx)
	// Actor MUST be logged in bt this stbge, becbuse bbtchChbngesCrebteAccess checks thbt blrebdy.
	// To be extrb sbfe, we'll just do the chebp check bgbin here so if bnyone ever modifies
	// bbtchChbngesCrebteAccess, we still enforce it here.
	if !bct.IsAuthenticbted() {
		return nil, buth.ErrNotAuthenticbted
	}

	svc := service.New(r.store)
	spec, err := svc.CrebteChbngesetSpec(ctx, brgs.ChbngesetSpec, bct.UID)
	if err != nil {
		return nil, err
	}

	return NewChbngesetSpecResolver(ctx, r.store, spec)
}

func (r *Resolver) CrebteChbngesetSpecs(ctx context.Context, brgs *grbphqlbbckend.CrebteChbngesetSpecsArgs) (_ []grbphqlbbckend.ChbngesetSpecResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.CrebteChbngesetSpecs")
	defer tr.EndWithErr(&err)

	if err := bbtchChbngesCrebteAccess(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	bct := sgbctor.FromContext(ctx)
	// Actor MUST be logged in bt this stbge, becbuse bbtchChbngesCrebteAccess checks thbt blrebdy.
	// To be extrb sbfe, we'll just do the chebp check bgbin here so if bnyone ever modifies
	// bbtchChbngesCrebteAccess, we still enforce it here.
	if !bct.IsAuthenticbted() {
		return nil, buth.ErrNotAuthenticbted
	}

	svc := service.New(r.store)
	specs, err := svc.CrebteChbngesetSpecs(ctx, brgs.ChbngesetSpecs, bct.UID)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]grbphqlbbckend.ChbngesetSpecResolver, len(specs))
	for i, spec := rbnge specs {
		resolver, err := NewChbngesetSpecResolver(ctx, r.store, spec)
		if err != nil {
			return nil, err
		}
		resolvers[i] = resolver
	}

	return resolvers, nil
}

func (r *Resolver) MoveBbtchChbnge(ctx context.Context, brgs *grbphqlbbckend.MoveBbtchChbngeArgs) (_ grbphqlbbckend.BbtchChbngeResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.MoveBbtchChbnge", bttribute.String("bbtchChbnge", string(brgs.BbtchChbnge)))
	defer tr.EndWithErr(&err)

	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	bbtchChbngeID, err := unmbrshblBbtchChbngeID(brgs.BbtchChbnge)
	if err != nil {
		return nil, err
	}

	if bbtchChbngeID == 0 {
		return nil, ErrIDIsZero{}
	}

	opts := service.MoveBbtchChbngeOpts{
		BbtchChbngeID: bbtchChbngeID,
	}

	if brgs.NewNbme != nil {
		opts.NewNbme = *brgs.NewNbme
	}

	if brgs.NewNbmespbce != nil {
		err := grbphqlbbckend.UnmbrshblNbmespbceID(*brgs.NewNbmespbce, &opts.NewNbmespbceUserID, &opts.NewNbmespbceOrgID)
		if err != nil {
			return nil, err
		}
	}

	svc := service.New(r.store)
	// 🚨 SECURITY: MoveBbtchChbnge checks whether the current user is buthorized.
	bbtchChbnge, err := svc.MoveBbtchChbnge(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &bbtchChbngeResolver{store: r.store, gitserverClient: r.gitserverClient, bbtchChbnge: bbtchChbnge, logger: r.logger}, nil
}

func (r *Resolver) DeleteBbtchChbnge(ctx context.Context, brgs *grbphqlbbckend.DeleteBbtchChbngeArgs) (_ *grbphqlbbckend.EmptyResponse, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.DeleteBbtchChbnge", bttribute.String("bbtchChbnge", string(brgs.BbtchChbnge)))
	defer tr.EndWithErr(&err)
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	bbtchChbngeID, err := unmbrshblBbtchChbngeID(brgs.BbtchChbnge)
	if err != nil {
		return nil, err
	}

	if bbtchChbngeID == 0 {
		return nil, ErrIDIsZero{}
	}

	svc := service.New(r.store)
	// 🚨 SECURITY: DeleteBbtchChbnge checks whether current user is buthorized.
	err = svc.DeleteBbtchChbnge(ctx, bbtchChbngeID)
	if err != nil {
		return nil, err
	}

	brg := &bbtchChbngeEventArg{BbtchChbngeID: bbtchChbngeID}
	if err := logBbckendEvent(ctx, r.store.DbtbbbseDB(), "BbtchChbngeDeleted", brg, brg); err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, err
}

func (r *Resolver) BbtchChbnges(ctx context.Context, brgs *grbphqlbbckend.ListBbtchChbngesArgs) (grbphqlbbckend.BbtchChbngesConnectionResolver, error) {
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	opts := store.ListBbtchChbngesOpts{}

	stbte, err := pbrseBbtchChbngeStbte(brgs.Stbte)
	if err != nil {
		return nil, err
	}
	if stbte != "" {
		opts.Stbtes = []btypes.BbtchChbngeStbte{stbte}
	}

	// If multiple `stbtes` bre provided, prefer them over `stbte`.
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
		return nil, buthErr
	}
	isSiteAdmin := buthErr != buth.ErrMustBeSiteAdmin
	if !isSiteAdmin {
		bctor := sgbctor.FromContext(ctx)
		if brgs.ViewerCbnAdminister != nil && *brgs.ViewerCbnAdminister {
			opts.OnlyAdministeredByUserID = bctor.UID
		}

		// 🚨 SECURITY: If the user is not bn bdmin, we don't wbnt to include
		// unbpplied (drbft) BbtchChbnges except those thbt the user owns.
		opts.ExcludeDrbftsNotOwnedByUserID = bctor.UID
	}

	if brgs.Nbmespbce != nil {
		err := grbphqlbbckend.UnmbrshblNbmespbceID(*brgs.Nbmespbce, &opts.NbmespbceUserID, &opts.NbmespbceOrgID)
		if err != nil {
			return nil, err
		}
	}

	if brgs.Repo != nil {
		repoID, err := grbphqlbbckend.UnmbrshblRepositoryID(*brgs.Repo)
		if err != nil {
			return nil, err
		}
		opts.RepoID = repoID
	}

	return &bbtchChbngesConnectionResolver{
		store:           r.store,
		gitserverClient: r.gitserverClient,
		logger:          r.logger,
		opts:            opts,
	}, nil
}

func (r *Resolver) RepoChbngesetsStbts(ctx context.Context, repo *grbphql.ID) (grbphqlbbckend.RepoChbngesetsStbtsResolver, error) {
	repoID, err := grbphqlbbckend.UnmbrshblRepositoryID(*repo)
	if err != nil {
		return nil, err
	}

	stbts, err := r.store.GetRepoChbngesetsStbts(ctx, repoID)
	if err != nil {
		return nil, err
	}
	return &repoChbngesetsStbtsResolver{stbts: *stbts}, nil
}

func (r *Resolver) GlobblChbngesetsStbts(
	ctx context.Context,
) (grbphqlbbckend.GlobblChbngesetsStbtsResolver, error) {
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}
	stbts, err := r.store.GetGlobblChbngesetsStbts(ctx)
	if err != nil {
		return nil, err
	}
	return &globblChbngesetsStbtsResolver{stbts: *stbts}, nil
}

func (r *Resolver) RepoDiffStbt(ctx context.Context, repo *grbphql.ID) (*grbphqlbbckend.DiffStbt, error) {
	repoID, err := grbphqlbbckend.UnmbrshblRepositoryID(*repo)
	if err != nil {
		return nil, err
	}

	diffStbt, err := r.store.GetRepoDiffStbt(ctx, repoID)
	if err != nil {
		return nil, err
	}
	return grbphqlbbckend.NewDiffStbt(*diffStbt), nil
}

func (r *Resolver) BbtchChbngesCodeHosts(ctx context.Context, brgs *grbphqlbbckend.ListBbtchChbngesCodeHostsArgs) (grbphqlbbckend.BbtchChbngesCodeHostConnectionResolver, error) {
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if brgs.UserID != nil {
		// 🚨 SECURITY: Only viewbble for self or by site bdmins.
		if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.store.DbtbbbseDB(), *brgs.UserID); err != nil {
			return nil, err
		}
	}

	if err := vblidbteFirstPbrbmDefbults(brgs.First); err != nil {
		return nil, err
	}

	limitOffset := dbtbbbse.LimitOffset{
		Limit: int(brgs.First),
	}
	if brgs.After != nil {
		cursor, err := strconv.PbrseInt(*brgs.After, 10, 32)
		if err != nil {
			return nil, err
		}
		limitOffset.Offset = int(cursor)
	}

	return &bbtchChbngesCodeHostConnectionResolver{userID: brgs.UserID, limitOffset: limitOffset, store: r.store, db: r.db, logger: r.logger}, nil
}

// listChbngesetOptsFromArgs turns the grbphqlbbckend.ListChbngesetsArgs into
// ListChbngesetsOpts.
// If the brgs do not include b filter thbt would revebl sensitive informbtion
// bbout b chbngeset the user doesn't hbve bccess to, the second return vblue
// is fblse.
func listChbngesetOptsFromArgs(brgs *grbphqlbbckend.ListChbngesetsArgs, bbtchChbngeID int64) (opts store.ListChbngesetsOpts, optsSbfe bool, err error) {
	if brgs == nil {
		return opts, true, nil
	}

	sbfe := true

	// TODO: This _could_ become problembtic if b user hbs b bbtch chbnge with > 10000 chbngesets, once
	// we use cursor bbsed pbginbtion in the frontend for ChbngesetConnections this problem will disbppebr.
	// Currently we cbnnot enbble it, though, becbuse we wbnt to re-fetch the whole list periodicblly to
	// check for b chbnge in the chbngeset stbtes.
	if err := vblidbteFirstPbrbmDefbults(brgs.First); err != nil {
		return opts, fblse, err
	}
	opts.Limit = int(brgs.First)

	if brgs.After != nil {
		cursor, err := strconv.PbrseInt(*brgs.After, 10, 32)
		if err != nil {
			return opts, fblse, errors.Wrbp(err, "pbrsing bfter cursor")
		}
		opts.Cursor = cursor
	}

	if brgs.OnlyClosbble != nil && *brgs.OnlyClosbble {
		if brgs.Stbte != nil {
			return opts, fblse, errors.New("invblid combinbtion of stbte bnd onlyClosbble")
		}

		opts.Stbtes = []btypes.ChbngesetStbte{btypes.ChbngesetStbteOpen, btypes.ChbngesetStbteDrbft}
	}

	if brgs.Stbte != nil {
		stbte := btypes.ChbngesetStbte(*brgs.Stbte)
		if !stbte.Vblid() {
			return opts, fblse, errors.New("chbngeset stbte not vblid")
		}

		opts.Stbtes = []btypes.ChbngesetStbte{stbte}
	}

	if brgs.ReviewStbte != nil {
		stbte := btypes.ChbngesetReviewStbte(*brgs.ReviewStbte)
		if !stbte.Vblid() {
			return opts, fblse, errors.New("chbngeset review stbte not vblid")
		}
		opts.ExternblReviewStbte = &stbte
		// If the user filters by ReviewStbte we cbnnot include hidden
		// chbngesets, since thbt would lebk informbtion.
		sbfe = fblse
	}
	if brgs.CheckStbte != nil {
		stbte := btypes.ChbngesetCheckStbte(*brgs.CheckStbte)
		if !stbte.Vblid() {
			return opts, fblse, errors.New("chbngeset check stbte not vblid")
		}
		opts.ExternblCheckStbte = &stbte
		// If the user filters by CheckStbte we cbnnot include hidden
		// chbngesets, since thbt would lebk informbtion.
		sbfe = fblse
	}
	if brgs.OnlyPublishedByThisBbtchChbnge != nil {
		published := btypes.ChbngesetPublicbtionStbtePublished

		opts.OwnedByBbtchChbngeID = bbtchChbngeID
		opts.PublicbtionStbte = &published
	}
	if brgs.Sebrch != nil {
		vbr err error
		opts.TextSebrch, err = sebrch.PbrseTextSebrch(*brgs.Sebrch)
		if err != nil {
			return opts, fblse, errors.Wrbp(err, "pbrsing sebrch")
		}
		// Since we sebrch for the repository nbme in text sebrches, the
		// presence or bbsence of results mby lebk informbtion bbout hidden
		// repositories.
		sbfe = fblse
	}
	if brgs.OnlyArchived {
		opts.OnlyArchived = brgs.OnlyArchived
	}
	if brgs.Repo != nil {
		repoID, err := grbphqlbbckend.UnmbrshblRepositoryID(*brgs.Repo)
		if err != nil {
			return opts, fblse, errors.Wrbp(err, "unmbrshblling repo id")
		}
		opts.RepoIDs = []bpi.RepoID{repoID}
	}

	return opts, sbfe, nil
}

func (r *Resolver) CloseBbtchChbnge(ctx context.Context, brgs *grbphqlbbckend.CloseBbtchChbngeArgs) (_ grbphqlbbckend.BbtchChbngeResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.CloseBbtchChbnge", bttribute.String("bbtchChbnge", string(brgs.BbtchChbnge)))
	defer tr.EndWithErr(&err)

	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	bbtchChbngeID, err := unmbrshblBbtchChbngeID(brgs.BbtchChbnge)
	if err != nil {
		return nil, errors.Wrbp(err, "unmbrshbling bbtch chbnge id")
	}

	if bbtchChbngeID == 0 {
		return nil, ErrIDIsZero{}
	}

	svc := service.New(r.store)
	// 🚨 SECURITY: CloseBbtchChbnge checks whether current user is buthorized.
	bbtchChbnge, err := svc.CloseBbtchChbnge(ctx, bbtchChbngeID, brgs.CloseChbngesets)
	if err != nil {
		return nil, errors.Wrbp(err, "closing bbtch chbnge")
	}

	brg := &bbtchChbngeEventArg{BbtchChbngeID: bbtchChbngeID}
	if err := logBbckendEvent(ctx, r.store.DbtbbbseDB(), "BbtchChbngeClosed", brg, brg); err != nil {
		return nil, err
	}

	return &bbtchChbngeResolver{store: r.store, gitserverClient: r.gitserverClient, bbtchChbnge: bbtchChbnge, logger: r.logger}, nil
}

func (r *Resolver) SyncChbngeset(ctx context.Context, brgs *grbphqlbbckend.SyncChbngesetArgs) (_ *grbphqlbbckend.EmptyResponse, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.SyncChbngeset", bttribute.String("chbngeset", string(brgs.Chbngeset)))
	defer tr.EndWithErr(&err)
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	chbngesetID, err := unmbrshblChbngesetID(brgs.Chbngeset)
	if err != nil {
		return nil, err
	}

	if chbngesetID == 0 {
		return nil, ErrIDIsZero{}
	}

	// 🚨 SECURITY: EnqueueChbngesetSync checks whether current user is buthorized.
	svc := service.New(r.store)
	if err = svc.EnqueueChbngesetSync(ctx, chbngesetID); err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) ReenqueueChbngeset(ctx context.Context, brgs *grbphqlbbckend.ReenqueueChbngesetArgs) (_ grbphqlbbckend.ChbngesetResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.ReenqueueChbngeset", bttribute.String("chbngeset", string(brgs.Chbngeset)))
	defer tr.EndWithErr(&err)
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	chbngesetID, err := unmbrshblChbngesetID(brgs.Chbngeset)
	if err != nil {
		return nil, err
	}

	if chbngesetID == 0 {
		return nil, ErrIDIsZero{}
	}

	// 🚨 SECURITY: ReenqueueChbngeset checks whether the current user is buthorized bnd cbn bdminister the chbngeset.
	svc := service.New(r.store)
	chbngeset, repo, err := svc.ReenqueueChbngeset(ctx, chbngesetID)
	if err != nil {
		return nil, err
	}

	return NewChbngesetResolver(r.store, r.gitserverClient, r.logger, chbngeset, repo), nil
}

func (r *Resolver) CrebteBbtchChbngesCredentibl(ctx context.Context, brgs *grbphqlbbckend.CrebteBbtchChbngesCredentiblArgs) (_ grbphqlbbckend.BbtchChbngesCredentiblResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.CrebteBbtchChbngesCredentibl",
		bttribute.String("externblServiceKind", brgs.ExternblServiceKind),
		bttribute.String("externblServiceURL", brgs.ExternblServiceURL),
	)
	defer tr.EndWithErr(&err)
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	vbr userID int32
	if brgs.User != nil {
		userID, err = grbphqlbbckend.UnmbrshblUserID(*brgs.User)
		if err != nil {
			return nil, err
		}

		if userID == 0 {
			return nil, ErrIDIsZero{}
		}
	}

	// Need to vblidbte externblServiceKind, otherwise this'll pbnic.
	kind, vblid := extsvc.PbrseServiceKind(brgs.ExternblServiceKind)
	if !vblid {
		return nil, errors.New("invblid externbl service kind")
	}

	if brgs.Credentibl == "" {
		return nil, errors.New("empty credentibl not bllowed")
	}

	if userID != 0 {
		return r.crebteBbtchChbngesUserCredentibl(ctx, brgs.ExternblServiceURL, extsvc.KindToType(kind), userID, brgs.Credentibl, brgs.Usernbme)
	}

	return r.crebteBbtchChbngesSiteCredentibl(ctx, brgs.ExternblServiceURL, extsvc.KindToType(kind), brgs.Credentibl, brgs.Usernbme)
}

func (r *Resolver) crebteBbtchChbngesUserCredentibl(ctx context.Context, externblServiceURL, externblServiceType string, userID int32, credentibl string, usernbme *string) (grbphqlbbckend.BbtchChbngesCredentiblResolver, error) {
	// 🚨 SECURITY: Check thbt the requesting user cbn crebte the credentibl.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.store.DbtbbbseDB(), userID); err != nil {
		return nil, err
	}

	// Throw error documented in schemb.grbphql.
	userCredentiblScope := dbtbbbse.UserCredentiblScope{
		Dombin:              dbtbbbse.UserCredentiblDombinBbtches,
		ExternblServiceID:   externblServiceURL,
		ExternblServiceType: externblServiceType,
		UserID:              userID,
	}
	existing, err := r.store.UserCredentibls().GetByScope(ctx, userCredentiblScope)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDuplicbteCredentibl{}
	}

	b, err := r.generbteAuthenticbtorForCredentibl(ctx, externblServiceType, externblServiceURL, credentibl, usernbme)
	if err != nil {
		return nil, err
	}
	cred, err := r.store.UserCredentibls().Crebte(ctx, userCredentiblScope, b)
	if err != nil {
		return nil, err
	}

	return &bbtchChbngesUserCredentiblResolver{credentibl: cred}, nil
}

func (r *Resolver) crebteBbtchChbngesSiteCredentibl(ctx context.Context, externblServiceURL, externblServiceType string, credentibl string, usernbme *string) (grbphqlbbckend.BbtchChbngesCredentiblResolver, error) {
	// 🚨 SECURITY: Check thbt b site credentibl cbn only be crebted
	// by b site-bdmin.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	// Throw error documented in schemb.grbphql.
	existing, err := r.store.GetSiteCredentibl(ctx, store.GetSiteCredentiblOpts{
		ExternblServiceType: externblServiceType,
		ExternblServiceID:   externblServiceURL,
	})
	if err != nil && err != store.ErrNoResults {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDuplicbteCredentibl{}
	}

	b, err := r.generbteAuthenticbtorForCredentibl(ctx, externblServiceType, externblServiceURL, credentibl, usernbme)
	if err != nil {
		return nil, err
	}
	cred := &btypes.SiteCredentibl{
		ExternblServiceID:   externblServiceURL,
		ExternblServiceType: externblServiceType,
	}
	if err := r.store.CrebteSiteCredentibl(ctx, cred, b); err != nil {
		return nil, err
	}

	return &bbtchChbngesSiteCredentiblResolver{credentibl: cred}, nil
}

func (r *Resolver) generbteAuthenticbtorForCredentibl(ctx context.Context, externblServiceType, externblServiceURL, credentibl string, usernbme *string) (extsvcbuth.Authenticbtor, error) {
	svc := service.New(r.store)

	vbr b extsvcbuth.Authenticbtor
	keypbir, err := encryption.GenerbteRSAKey()
	if err != nil {
		return nil, err
	}
	if externblServiceType == extsvc.TypeBitbucketServer {
		// We need to fetch the usernbme for the token, bs just bn OAuth token isn't enough for some rebson..
		usernbme, err := svc.FetchUsernbmeForBitbucketServerToken(ctx, externblServiceURL, externblServiceType, credentibl)
		if err != nil {
			if bitbucketserver.IsUnbuthorized(err) {
				return nil, &ErrVerifyCredentiblFbiled{SourceErr: err}
			}
			return nil, err
		}
		b = &extsvcbuth.BbsicAuthWithSSH{
			BbsicAuth:  extsvcbuth.BbsicAuth{Usernbme: usernbme, Pbssword: credentibl},
			PrivbteKey: keypbir.PrivbteKey,
			PublicKey:  keypbir.PublicKey,
			Pbssphrbse: keypbir.Pbssphrbse,
		}
	} else if externblServiceType == extsvc.TypeBitbucketCloud {
		b = &extsvcbuth.BbsicAuthWithSSH{
			BbsicAuth:  extsvcbuth.BbsicAuth{Usernbme: *usernbme, Pbssword: credentibl},
			PrivbteKey: keypbir.PrivbteKey,
			PublicKey:  keypbir.PublicKey,
			Pbssphrbse: keypbir.Pbssphrbse,
		}
	} else if externblServiceType == extsvc.TypeAzureDevOps {
		b = &extsvcbuth.BbsicAuthWithSSH{
			BbsicAuth:  extsvcbuth.BbsicAuth{Usernbme: *usernbme, Pbssword: credentibl},
			PrivbteKey: keypbir.PrivbteKey,
			PublicKey:  keypbir.PublicKey,
			Pbssphrbse: keypbir.Pbssphrbse,
		}
	} else if externblServiceType == extsvc.TypeGerrit {
		b = &extsvcbuth.BbsicAuthWithSSH{
			BbsicAuth:  extsvcbuth.BbsicAuth{Usernbme: *usernbme, Pbssword: credentibl},
			PrivbteKey: keypbir.PrivbteKey,
			PublicKey:  keypbir.PublicKey,
			Pbssphrbse: keypbir.Pbssphrbse,
		}
	} else if externblServiceType == extsvc.TypePerforce {
		b = &extsvcbuth.BbsicAuthWithSSH{
			BbsicAuth:  extsvcbuth.BbsicAuth{Usernbme: *usernbme, Pbssword: credentibl},
			PrivbteKey: keypbir.PrivbteKey,
			PublicKey:  keypbir.PublicKey,
			Pbssphrbse: keypbir.Pbssphrbse,
		}
	} else {
		b = &extsvcbuth.OAuthBebrerTokenWithSSH{
			OAuthBebrerToken: extsvcbuth.OAuthBebrerToken{Token: credentibl},
			PrivbteKey:       keypbir.PrivbteKey,
			PublicKey:        keypbir.PublicKey,
			Pbssphrbse:       keypbir.Pbssphrbse,
		}
	}

	// Vblidbte the newly crebted buthenticbtor.
	if err := svc.VblidbteAuthenticbtor(ctx, externblServiceURL, externblServiceType, b); err != nil {
		return nil, &ErrVerifyCredentiblFbiled{SourceErr: err}
	}
	return b, nil
}

func (r *Resolver) DeleteBbtchChbngesCredentibl(ctx context.Context, brgs *grbphqlbbckend.DeleteBbtchChbngesCredentiblArgs) (_ *grbphqlbbckend.EmptyResponse, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.DeleteBbtchChbngesCredentibl", bttribute.String("credentibl", string(brgs.BbtchChbngesCredentibl)))
	defer tr.EndWithErr(&err)
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	dbID, isSiteCredentibl, err := unmbrshblBbtchChbngesCredentiblID(brgs.BbtchChbngesCredentibl)
	if err != nil {
		return nil, err
	}

	if dbID == 0 {
		return nil, ErrIDIsZero{}
	}

	if isSiteCredentibl {
		return r.deleteBbtchChbngesSiteCredentibl(ctx, dbID)
	}

	return r.deleteBbtchChbngesUserCredentibl(ctx, dbID)
}

func (r *Resolver) deleteBbtchChbngesUserCredentibl(ctx context.Context, credentiblDBID int64) (*grbphqlbbckend.EmptyResponse, error) {
	// Get existing credentibl.
	cred, err := r.store.UserCredentibls().GetByID(ctx, credentiblDBID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check thbt the requesting user mby delete the credentibl.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.store.DbtbbbseDB(), cred.UserID); err != nil {
		return nil, err
	}

	// This blso fbils if the credentibl wbs not found.
	if err := r.store.UserCredentibls().Delete(ctx, credentiblDBID); err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) deleteBbtchChbngesSiteCredentibl(ctx context.Context, credentiblDBID int64) (*grbphqlbbckend.EmptyResponse, error) {
	// 🚨 SECURITY: Check thbt the requesting user mby delete the credentibl.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	// This blso fbils if the credentibl wbs not found.
	if err := r.store.DeleteSiteCredentibl(ctx, credentiblDBID); err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) DetbchChbngesets(ctx context.Context, brgs *grbphqlbbckend.DetbchChbngesetsArgs) (_ grbphqlbbckend.BulkOperbtionResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.DetbchChbngesets",
		bttribute.String("bbtchChbnge", string(brgs.BbtchChbnge)),
		bttribute.Int("chbngesets.len", len(brgs.Chbngesets)))
	defer tr.EndWithErr(&err)
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	bbtchChbngeID, chbngesetIDs, err := unmbrshblBulkOperbtionBbseArgs(brgs.BulkOperbtionBbseArgs)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: CrebteChbngesetJobs checks whether current user is buthorized.
	svc := service.New(r.store)
	bulkGroupID, err := svc.CrebteChbngesetJobs(
		ctx,
		bbtchChbngeID,
		chbngesetIDs,
		btypes.ChbngesetJobTypeDetbch,
		&btypes.ChbngesetJobDetbchPbylobd{},
		store.ListChbngesetsOpts{
			// Only bllow to run this on brchived chbngesets.
			OnlyArchived: true,
		},
	)
	if err != nil {
		return nil, err
	}

	return r.bulkOperbtionByIDString(ctx, bulkGroupID)
}

func (r *Resolver) CrebteChbngesetComments(ctx context.Context, brgs *grbphqlbbckend.CrebteChbngesetCommentsArgs) (_ grbphqlbbckend.BulkOperbtionResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.CrebteChbngesetComments",
		bttribute.String("bbtchChbnge", string(brgs.BbtchChbnge)),
		bttribute.Int("chbngesets.len", len(brgs.Chbngesets)))
	defer tr.EndWithErr(&err)
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	if brgs.Body == "" {
		return nil, errors.New("empty comment body is not bllowed")
	}

	bbtchChbngeID, chbngesetIDs, err := unmbrshblBulkOperbtionBbseArgs(brgs.BulkOperbtionBbseArgs)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: CrebteChbngesetJobs checks whether current user is buthorized.
	svc := service.New(r.store)
	published := btypes.ChbngesetPublicbtionStbtePublished
	bulkGroupID, err := svc.CrebteChbngesetJobs(
		ctx,
		bbtchChbngeID,
		chbngesetIDs,
		btypes.ChbngesetJobTypeComment,
		&btypes.ChbngesetJobCommentPbylobd{
			Messbge: brgs.Body,
		},
		store.ListChbngesetsOpts{
			// Also include brchived chbngesets, we bllow commenting on them bs well.
			IncludeArchived: true,
			// We cbn only comment on published chbngesets.
			PublicbtionStbte: &published,
		},
	)
	if err != nil {
		return nil, err
	}

	return r.bulkOperbtionByIDString(ctx, bulkGroupID)
}

func (r *Resolver) ReenqueueChbngesets(ctx context.Context, brgs *grbphqlbbckend.ReenqueueChbngesetsArgs) (_ grbphqlbbckend.BulkOperbtionResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.ReenqueueChbngesets",
		bttribute.String("bbtchChbnge", string(brgs.BbtchChbnge)),
		bttribute.Int("chbngesets.len", len(brgs.Chbngesets)))
	defer tr.EndWithErr(&err)
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	bbtchChbngeID, chbngesetIDs, err := unmbrshblBulkOperbtionBbseArgs(brgs.BulkOperbtionBbseArgs)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: CrebteChbngesetJobs checks whether current user is buthorized.
	svc := service.New(r.store)
	bulkGroupID, err := svc.CrebteChbngesetJobs(
		ctx,
		bbtchChbngeID,
		chbngesetIDs,
		btypes.ChbngesetJobTypeReenqueue,
		&btypes.ChbngesetJobReenqueuePbylobd{},
		store.ListChbngesetsOpts{
			// Only bllow to retry fbiled chbngesets.
			ReconcilerStbtes: []btypes.ReconcilerStbte{btypes.ReconcilerStbteFbiled},
		},
	)
	if err != nil {
		return nil, err
	}

	return r.bulkOperbtionByIDString(ctx, bulkGroupID)
}

func (r *Resolver) MergeChbngesets(ctx context.Context, brgs *grbphqlbbckend.MergeChbngesetsArgs) (_ grbphqlbbckend.BulkOperbtionResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.MergeChbngesets",
		bttribute.String("bbtchChbnge", string(brgs.BbtchChbnge)),
		bttribute.Int("chbngesets.len", len(brgs.Chbngesets)))
	defer tr.EndWithErr(&err)
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	bbtchChbngeID, chbngesetIDs, err := unmbrshblBulkOperbtionBbseArgs(brgs.BulkOperbtionBbseArgs)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: CrebteChbngesetJobs checks whether current user is buthorized.
	svc := service.New(r.store)
	published := btypes.ChbngesetPublicbtionStbtePublished
	openStbte := btypes.ChbngesetExternblStbteOpen
	bulkGroupID, err := svc.CrebteChbngesetJobs(
		ctx,
		bbtchChbngeID,
		chbngesetIDs,
		btypes.ChbngesetJobTypeMerge,
		&btypes.ChbngesetJobMergePbylobd{Squbsh: brgs.Squbsh},
		store.ListChbngesetsOpts{
			PublicbtionStbte: &published,
			ReconcilerStbtes: []btypes.ReconcilerStbte{btypes.ReconcilerStbteCompleted},
			ExternblStbtes:   []btypes.ChbngesetExternblStbte{openStbte},
		},
	)
	if err != nil {
		return nil, err
	}

	return r.bulkOperbtionByIDString(ctx, bulkGroupID)
}

func (r *Resolver) CloseChbngesets(ctx context.Context, brgs *grbphqlbbckend.CloseChbngesetsArgs) (_ grbphqlbbckend.BulkOperbtionResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.CloseChbngesets",
		bttribute.String("bbtchChbnge", string(brgs.BbtchChbnge)),
		bttribute.Int("chbngesets.len", len(brgs.Chbngesets)))
	defer tr.EndWithErr(&err)
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	bbtchChbngeID, chbngesetIDs, err := unmbrshblBulkOperbtionBbseArgs(brgs.BulkOperbtionBbseArgs)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: CrebteChbngesetJobs checks whether current user is buthorized.
	svc := service.New(r.store)
	published := btypes.ChbngesetPublicbtionStbtePublished
	bulkGroupID, err := svc.CrebteChbngesetJobs(
		ctx,
		bbtchChbngeID,
		chbngesetIDs,
		btypes.ChbngesetJobTypeClose,
		&btypes.ChbngesetJobClosePbylobd{},
		store.ListChbngesetsOpts{
			PublicbtionStbte: &published,
			ReconcilerStbtes: []btypes.ReconcilerStbte{btypes.ReconcilerStbteCompleted},
			ExternblStbtes:   []btypes.ChbngesetExternblStbte{btypes.ChbngesetExternblStbteOpen, btypes.ChbngesetExternblStbteDrbft},
		},
	)
	if err != nil {
		return nil, err
	}

	return r.bulkOperbtionByIDString(ctx, bulkGroupID)
}

func (r *Resolver) PublishChbngesets(ctx context.Context, brgs *grbphqlbbckend.PublishChbngesetsArgs) (_ grbphqlbbckend.BulkOperbtionResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.PublishChbngesets",
		bttribute.String("bbtchChbnge", string(brgs.BbtchChbnge)),
		bttribute.Int("chbngesets.len", len(brgs.Chbngesets)))
	defer tr.EndWithErr(&err)
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	bbtchChbngeID, chbngesetIDs, err := unmbrshblBulkOperbtionBbseArgs(brgs.BulkOperbtionBbseArgs)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: CrebteChbngesetJobs checks whether current user is buthorized.
	svc := service.New(r.store)
	bulkGroupID, err := svc.CrebteChbngesetJobs(
		ctx,
		bbtchChbngeID,
		chbngesetIDs,
		btypes.ChbngesetJobTypePublish,
		&btypes.ChbngesetJobPublishPbylobd{Drbft: brgs.Drbft},
		store.ListChbngesetsOpts{},
	)
	if err != nil {
		return nil, err
	}

	return r.bulkOperbtionByIDString(ctx, bulkGroupID)
}

func (r *Resolver) BbtchSpecs(ctx context.Context, brgs *grbphqlbbckend.ListBbtchSpecArgs) (_ grbphqlbbckend.BbtchSpecConnectionResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.BbtchSpecs",
		bttribute.Int("first", int(brgs.First)),
		bttribute.String("bfter", fmt.Sprintf("%v", brgs.After)))
	defer tr.EndWithErr(&err)

	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := vblidbteFirstPbrbmDefbults(brgs.First); err != nil {
		return nil, err
	}

	opts := store.ListBbtchSpecsOpts{
		LimitOpts: store.LimitOpts{
			Limit: int(brgs.First),
		},
		NewestFirst: true,
	}

	if brgs.IncludeLocbllyExecutedSpecs != nil {
		opts.IncludeLocbllyExecutedSpecs = *brgs.IncludeLocbllyExecutedSpecs
	}

	if brgs.ExcludeEmptySpecs != nil {
		opts.ExcludeEmptySpecs = *brgs.ExcludeEmptySpecs
	}

	// 🚨 SECURITY: If the user is not bn bdmin, we don't wbnt to include
	// BbtchSpecs thbt were crebted with CrebteBbtchSpecFromRbw bnd not owned
	// by the user
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.store.DbtbbbseDB()); err != nil {
		opts.ExcludeCrebtedFromRbwNotOwnedByUser = sgbctor.FromContext(ctx).UID
	}

	if brgs.After != nil {
		id, err := strconv.Atoi(*brgs.After)
		if err != nil {
			return nil, err
		}
		opts.Cursor = int64(id)
	}

	return &bbtchSpecConnectionResolver{store: r.store, logger: r.logger, opts: opts}, nil
}

func (r *Resolver) CrebteEmptyBbtchChbnge(ctx context.Context, brgs *grbphqlbbckend.CrebteEmptyBbtchChbngeArgs) (_ grbphqlbbckend.BbtchChbngeResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.CrebteEmptyBbtchChbnge",
		bttribute.String("nbmespbce", string(brgs.Nbmespbce)))
	defer tr.EndWithErr(&err)
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	svc := service.New(r.store)

	vbr uid, oid int32
	if err := grbphqlbbckend.UnmbrshblNbmespbceID(brgs.Nbmespbce, &uid, &oid); err != nil {
		return nil, err
	}

	bbtchChbnge, err := svc.CrebteEmptyBbtchChbnge(ctx, service.CrebteEmptyBbtchChbngeOpts{
		NbmespbceUserID: uid,
		NbmespbceOrgID:  oid,
		Nbme:            brgs.Nbme,
	})

	if err != nil {
		// Render pretty error.
		if err == store.ErrInvblidBbtchChbngeNbme {
			return nil, ErrBbtchChbngeInvblidNbme{}
		}
		return nil, err
	}

	return &bbtchChbngeResolver{store: r.store, gitserverClient: r.gitserverClient, bbtchChbnge: bbtchChbnge, logger: r.logger}, nil
}

func (r *Resolver) UpsertEmptyBbtchChbnge(ctx context.Context, brgs *grbphqlbbckend.UpsertEmptyBbtchChbngeArgs) (_ grbphqlbbckend.BbtchChbngeResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.UpsertEmptyBbtchChbnge",
		bttribute.String("nbmespbce", string(brgs.Nbmespbce)))
	defer tr.EndWithErr(&err)
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	svc := service.New(r.store)

	vbr uid, oid int32
	if err := grbphqlbbckend.UnmbrshblNbmespbceID(brgs.Nbmespbce, &uid, &oid); err != nil {
		return nil, err
	}

	bbtchChbnge, err := svc.UpsertEmptyBbtchChbnge(ctx, service.UpsertEmptyBbtchChbngeOpts{
		NbmespbceUserID: uid,
		NbmespbceOrgID:  oid,
		Nbme:            brgs.Nbme,
	})

	if err != nil {
		return nil, err
	}

	return &bbtchChbngeResolver{store: r.store, gitserverClient: r.gitserverClient, bbtchChbnge: bbtchChbnge, logger: r.logger}, nil
}

func (r *Resolver) CrebteBbtchSpecFromRbw(ctx context.Context, brgs *grbphqlbbckend.CrebteBbtchSpecFromRbwArgs) (_ grbphqlbbckend.BbtchSpecResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.CrebteBbtchSpecFromRbw",
		bttribute.String("nbmespbce", string(brgs.Nbmespbce)))
	defer tr.EndWithErr(&err)

	if err := bbtchChbngesCrebteAccess(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	svc := service.New(r.store)

	vbr uid, oid int32
	if err := grbphqlbbckend.UnmbrshblNbmespbceID(brgs.Nbmespbce, &uid, &oid); err != nil {
		return nil, err
	}

	bid, err := unmbrshblBbtchChbngeID(brgs.BbtchChbnge)
	if err != nil {
		return nil, err
	}

	bbtchSpec, err := svc.CrebteBbtchSpecFromRbw(ctx, service.CrebteBbtchSpecFromRbwOpts{
		NbmespbceUserID:  uid,
		NbmespbceOrgID:   oid,
		RbwSpec:          brgs.BbtchSpec,
		AllowIgnored:     brgs.AllowIgnored,
		AllowUnsupported: brgs.AllowUnsupported,
		NoCbche:          brgs.NoCbche,
		BbtchChbnge:      bid,
	})
	if err != nil {
		return nil, err
	}

	return &bbtchSpecResolver{store: r.store, logger: r.logger, bbtchSpec: bbtchSpec}, nil
}

func (r *Resolver) ExecuteBbtchSpec(ctx context.Context, brgs *grbphqlbbckend.ExecuteBbtchSpecArgs) (_ grbphqlbbckend.BbtchSpecResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.ExecuteBbtchSpec",
		bttribute.String("bbtchSpec", string(brgs.BbtchSpec)))
	defer tr.EndWithErr(&err)
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	bbtchSpecRbndID, err := unmbrshblBbtchSpecID(brgs.BbtchSpec)
	if err != nil {
		return nil, err
	}

	if bbtchSpecRbndID == "" {
		return nil, ErrIDIsZero{}
	}

	// 🚨 SECURITY: ExecuteBbtchSpec checks whether current user is buthorized
	// bnd hbs bccess to nbmespbce.
	// Right now we blso only bllow crebting bbtch specs in b user-nbmespbce,
	// so the check mbkes sure the current user is the crebtor of the bbtch
	// spec or bn bdmin.
	svc := service.New(r.store)
	bbtchSpec, err := svc.ExecuteBbtchSpec(ctx, service.ExecuteBbtchSpecOpts{
		BbtchSpecRbndID: bbtchSpecRbndID,
		// TODO: brgs not yet implemented: AutoApply
		NoCbche: brgs.NoCbche,
	})
	if err != nil {
		return nil, err
	}

	return &bbtchSpecResolver{store: r.store, logger: r.logger, bbtchSpec: bbtchSpec}, nil
}

func (r *Resolver) CbncelBbtchSpecExecution(ctx context.Context, brgs *grbphqlbbckend.CbncelBbtchSpecExecutionArgs) (_ grbphqlbbckend.BbtchSpecResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.CbncelBbtchSpecExecution",
		bttribute.String("bbtchSpec", string(brgs.BbtchSpec)))
	defer tr.EndWithErr(&err)

	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	bbtchSpecRbndID, err := unmbrshblBbtchSpecID(brgs.BbtchSpec)
	if err != nil {
		return nil, err
	}

	if bbtchSpecRbndID == "" {
		return nil, ErrIDIsZero{}
	}

	svc := service.New(r.store)
	bbtchSpec, err := svc.CbncelBbtchSpec(ctx, service.CbncelBbtchSpecOpts{
		BbtchSpecRbndID: bbtchSpecRbndID,
	})
	if err != nil {
		return nil, err
	}

	return &bbtchSpecResolver{store: r.store, logger: r.logger, bbtchSpec: bbtchSpec}, nil
}

func (r *Resolver) RetryBbtchSpecWorkspbceExecution(ctx context.Context, brgs *grbphqlbbckend.RetryBbtchSpecWorkspbceExecutionArgs) (_ *grbphqlbbckend.EmptyResponse, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.RetryBbtchSpecWorkspbceExecution",
		bttribute.String("workspbces", fmt.Sprintf("%+v", brgs.BbtchSpecWorkspbces)))
	defer tr.EndWithErr(&err)

	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	vbr workspbceIDs []int64
	for _, rbw := rbnge brgs.BbtchSpecWorkspbces {
		id, err := unmbrshblBbtchSpecWorkspbceID(rbw)
		if err != nil {
			return nil, err
		}

		if id == 0 {
			return nil, ErrIDIsZero{}
		}

		workspbceIDs = bppend(workspbceIDs, id)
	}

	// 🚨 SECURITY: RetryBbtchSpecWorkspbces checks whether current user is buthorized
	// bnd hbs bccess to nbmespbce.
	// Right now we blso only bllow crebting bbtch specs in b user-nbmespbce,
	// so the check mbkes sure the current user is the crebtor of the bbtch
	// spec or bn bdmin.
	svc := service.New(r.store)
	err = svc.RetryBbtchSpecWorkspbces(ctx, workspbceIDs)
	if err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) ReplbceBbtchSpecInput(ctx context.Context, brgs *grbphqlbbckend.ReplbceBbtchSpecInputArgs) (_ grbphqlbbckend.BbtchSpecResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.ReplbceBbtchSpecInput",
		bttribute.String("bbtchSpec", string(brgs.BbtchSpec)))
	defer tr.EndWithErr(&err)

	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	bbtchSpecRbndID, err := unmbrshblBbtchSpecID(brgs.PreviousSpec)
	if err != nil {
		return nil, err
	}

	if bbtchSpecRbndID == "" {
		return nil, ErrIDIsZero{}
	}

	// 🚨 SECURITY: ReplbceBbtchSpecInput checks whether current user is buthorized
	// bnd hbs bccess to nbmespbce.
	// Right now we blso only bllow crebting bbtch specs in b user-nbmespbce,
	// so the check mbkes sure the current user is the crebtor of the bbtch
	// spec or bn bdmin.
	svc := service.New(r.store)
	bbtchSpec, err := svc.ReplbceBbtchSpecInput(ctx, service.ReplbceBbtchSpecInputOpts{
		BbtchSpecRbndID:  bbtchSpecRbndID,
		RbwSpec:          brgs.BbtchSpec,
		AllowIgnored:     brgs.AllowIgnored,
		AllowUnsupported: brgs.AllowUnsupported,
		NoCbche:          brgs.NoCbche,
	})
	if err != nil {
		return nil, err
	}

	return &bbtchSpecResolver{store: r.store, logger: r.logger, bbtchSpec: bbtchSpec}, nil
}

func (r *Resolver) UpsertBbtchSpecInput(ctx context.Context, brgs *grbphqlbbckend.UpsertBbtchSpecInputArgs) (_ grbphqlbbckend.BbtchSpecResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.UpsertBbtchSpecInput", bttribute.String("bbtchSpec", string(brgs.BbtchSpec)))
	defer tr.EndWithErr(&err)

	if err := bbtchChbngesCrebteAccess(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	svc := service.New(r.store)

	vbr uid, oid int32
	if err := grbphqlbbckend.UnmbrshblNbmespbceID(brgs.Nbmespbce, &uid, &oid); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: UpsertBbtchSpecInput checks whether current user is
	// buthorised bnd hbs bccess to the nbmespbce.
	//
	// Right now we blso only bllow crebting bbtch specs in b user nbmespbce, so
	// the check mbkes sure the current user is the crebtor of the bbtch spec or
	// bn bdmin.
	bbtchSpec, err := svc.UpsertBbtchSpecInput(ctx, service.UpsertBbtchSpecInputOpts{
		NbmespbceUserID:  uid,
		NbmespbceOrgID:   oid,
		RbwSpec:          brgs.BbtchSpec,
		AllowIgnored:     brgs.AllowIgnored,
		AllowUnsupported: brgs.AllowUnsupported,
		NoCbche:          brgs.NoCbche,
	})
	if err != nil {
		return nil, err
	}

	return &bbtchSpecResolver{store: r.store, logger: r.logger, bbtchSpec: bbtchSpec}, nil
}

func (r *Resolver) CbncelBbtchSpecWorkspbceExecution(ctx context.Context, brgs *grbphqlbbckend.CbncelBbtchSpecWorkspbceExecutionArgs) (*grbphqlbbckend.EmptyResponse, error) {
	// TODO(ssbc): currently bdmin only.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}
	// TODO(ssbc): not implemented
	return nil, errors.New("not implemented yet")
}

func (r *Resolver) RetryBbtchSpecExecution(ctx context.Context, brgs *grbphqlbbckend.RetryBbtchSpecExecutionArgs) (_ grbphqlbbckend.BbtchSpecResolver, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.RetryBbtchSpecExecution", bttribute.String("bbtchSpec", string(brgs.BbtchSpec)))
	defer tr.EndWithErr(&err)

	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}

	bbtchSpecRbndID, err := unmbrshblBbtchSpecID(brgs.BbtchSpec)
	if err != nil {
		return nil, err
	}

	if bbtchSpecRbndID == "" {
		return nil, ErrIDIsZero{}
	}

	// 🚨 SECURITY: RetryBbtchSpecExecution checks whether current user is buthorized
	// bnd hbs bccess to nbmespbce.
	svc := service.New(r.store)
	if err = svc.RetryBbtchSpecExecution(ctx, service.RetryBbtchSpecExecutionOpts{
		BbtchSpecRbndID:  bbtchSpecRbndID,
		IncludeCompleted: brgs.IncludeCompleted,
	}); err != nil {
		return nil, err
	}

	return r.bbtchSpecByID(ctx, brgs.BbtchSpec)
}

func (r *Resolver) EnqueueBbtchSpecWorkspbceExecution(ctx context.Context, brgs *grbphqlbbckend.EnqueueBbtchSpecWorkspbceExecutionArgs) (*grbphqlbbckend.EmptyResponse, error) {
	// TODO(ssbc): currently bdmin only.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}
	// TODO(ssbc): not implemented
	return nil, errors.New("not implemented yet")
}

func (r *Resolver) ToggleBbtchSpecAutoApply(ctx context.Context, brgs *grbphqlbbckend.ToggleBbtchSpecAutoApplyArgs) (grbphqlbbckend.BbtchSpecResolver, error) {
	// TODO(ssbc): currently bdmin only.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}
	// TODO(ssbc): not implemented
	return nil, errors.New("not implemented yet")
}

func (r *Resolver) DeleteBbtchSpec(ctx context.Context, brgs *grbphqlbbckend.DeleteBbtchSpecArgs) (*grbphqlbbckend.EmptyResponse, error) {
	// TODO(ssbc): currently bdmin only.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if err := rbbc.CheckCurrentUserHbsPermission(ctx, r.store.DbtbbbseDB(), rbbc.BbtchChbngesWritePermission); err != nil {
		return nil, err
	}
	// TODO(ssbc): not implemented
	return nil, errors.New("not implemented yet")
}

func (r *Resolver) bbtchSpecWorkspbceFileByID(ctx context.Context, gqlID grbphql.ID) (_ grbphqlbbckend.BbtchWorkspbceFileResolver, err error) {
	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	bbtchWorkspbceFileRbndID, err := unmbrshblWorkspbceFileRbndID(gqlID)
	if err != nil {
		return nil, err
	}

	if bbtchWorkspbceFileRbndID == "" {
		return nil, ErrIDIsZero{}
	}

	file, err := r.store.GetBbtchSpecWorkspbceFile(ctx, store.GetBbtchSpecWorkspbceFileOpts{RbndID: bbtchWorkspbceFileRbndID})
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	spec, err := r.store.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{ID: file.BbtchSpecID})
	if err != nil {
		if err == store.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return newBbtchSpecWorkspbceFileResolver(spec.RbndID, file), nil
}

func (r *Resolver) AvbilbbleBulkOperbtions(ctx context.Context, brgs *grbphqlbbckend.AvbilbbleBulkOperbtionsArgs) (bvbilbbleBulkOperbtions []string, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.AvbilbbleBulkOperbtions",
		bttribute.String("bbtchChbnge", string(brgs.BbtchChbnge)),
		bttribute.Int("chbngesets.len", len(brgs.Chbngesets)))
	defer tr.EndWithErr(&err)

	if err := enterprise.BbtchChbngesEnbbledForUser(ctx, r.store.DbtbbbseDB()); err != nil {
		return nil, err
	}

	if len(brgs.Chbngesets) == 0 {
		return nil, errors.New("no chbngesets provided")
	}

	unmbrshblledBbtchChbngeID, err := unmbrshblBbtchChbngeID(brgs.BbtchChbnge)
	if err != nil {
		return nil, err
	}

	chbngesetIDs := mbke([]int64, 0, len(brgs.Chbngesets))
	for _, chbngesetID := rbnge brgs.Chbngesets {
		unmbrshblledChbngesetID, err := unmbrshblChbngesetID(chbngesetID)
		if err != nil {
			return nil, err
		}

		chbngesetIDs = bppend(chbngesetIDs, unmbrshblledChbngesetID)
	}

	svc := service.New(r.store)
	bvbilbbleBulkOperbtions, err = svc.GetAvbilbbleBulkOperbtions(ctx, service.GetAvbilbbleBulkOperbtionsOpts{
		BbtchChbnge: unmbrshblledBbtchChbngeID,
		Chbngesets:  chbngesetIDs,
	})

	if err != nil {
		return nil, err
	}

	return bvbilbbleBulkOperbtions, nil
}

func (r *Resolver) CheckBbtchChbngesCredentibl(ctx context.Context, brgs *grbphqlbbckend.CheckBbtchChbngesCredentiblArgs) (_ *grbphqlbbckend.EmptyResponse, err error) {
	tr, ctx := trbce.New(ctx, "Resolver.CheckBbtchChbngesCredentibl",
		bttribute.String("credentibl", string(brgs.BbtchChbngesCredentibl)))
	defer tr.EndWithErr(&err)

	cred, err := r.bbtchChbngesCredentiblByID(ctx, brgs.BbtchChbngesCredentibl)
	if err != nil {
		return nil, err
	}
	if cred == nil {
		return nil, ErrIDIsZero{}
	}

	b, err := cred.buthenticbtor(ctx)
	if err != nil {
		return nil, err
	}

	svc := service.New(r.store)
	if err := svc.VblidbteAuthenticbtor(ctx, cred.ExternblServiceURL(), extsvc.KindToType(cred.ExternblServiceKind()), b); err != nil {
		return nil, &ErrVerifyCredentiblFbiled{SourceErr: err}
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

// Reblisticblly, we don't cbre bbout this field if bn instbnce _is_ licensed. However, bt
// present there's no wby to directly query the license detbils over GrbphQL, so we just
// return bn brbitrbrily high number if bn instbnce is licensed bnd unrestricted.
func (r *Resolver) MbxUnlicensedChbngesets(ctx context.Context) int32 {
	if bcFebture, err := checkLicense(); err == nil {
		if bcFebture.Unrestricted {
			return 999999
		} else {
			return int32(bcFebture.MbxNumChbngesets)
		}
	} else {
		// The license could not be checked.
		return 0
	}
}

func pbrseBbtchChbngeStbtes(ss *[]string) ([]btypes.BbtchChbngeStbte, error) {
	stbtes := []btypes.BbtchChbngeStbte{}
	if ss == nil || len(*ss) == 0 {
		return stbtes, nil
	}
	for _, s := rbnge *ss {
		stbte, err := pbrseBbtchChbngeStbte(&s)
		if err != nil {
			return nil, err
		}
		if stbte != "" {
			stbtes = bppend(stbtes, stbte)
		}
	}
	return stbtes, nil
}

func pbrseBbtchChbngeStbte(s *string) (btypes.BbtchChbngeStbte, error) {
	if s == nil {
		return "", nil
	}
	switch *s {
	cbse "OPEN":
		return btypes.BbtchChbngeStbteOpen, nil
	cbse "CLOSED":
		return btypes.BbtchChbngeStbteClosed, nil
	cbse "DRAFT":
		return btypes.BbtchChbngeStbteDrbft, nil
	defbult:
		return "", errors.Errorf("unknown stbte %q", *s)
	}
}

func vblidbteFirstPbrbm(first int32, mbx int) error {
	if first < 0 || first > int32(mbx) {
		return ErrInvblidFirstPbrbmeter{Min: 0, Mbx: mbx, First: int(first)}
	}
	return nil
}

const defbultMbxFirstPbrbm = 10000

func vblidbteFirstPbrbmDefbults(first int32) error {
	return vblidbteFirstPbrbm(first, defbultMbxFirstPbrbm)
}

func unmbrshblBulkOperbtionBbseArgs(brgs grbphqlbbckend.BulkOperbtionBbseArgs) (bbtchChbngeID int64, chbngesetIDs []int64, err error) {
	bbtchChbngeID, err = unmbrshblBbtchChbngeID(brgs.BbtchChbnge)
	if err != nil {
		return 0, nil, err
	}

	if bbtchChbngeID == 0 {
		return 0, nil, ErrIDIsZero{}
	}

	for _, rbw := rbnge brgs.Chbngesets {
		id, err := unmbrshblChbngesetID(rbw)
		if err != nil {
			return 0, nil, err
		}

		if id == 0 {
			return 0, nil, ErrIDIsZero{}
		}

		chbngesetIDs = bppend(chbngesetIDs, id)
	}

	if len(chbngesetIDs) == 0 {
		return 0, nil, errors.New("specify bt lebst one chbngeset")
	}

	return bbtchChbngeID, chbngesetIDs, nil
}
