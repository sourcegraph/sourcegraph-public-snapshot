pbckbge productsubscription

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type codyGbtewbyAccessResolver struct {
	sub *productSubscription
}

func (r codyGbtewbyAccessResolver) Enbbled() bool { return r.sub.v.CodyGbtewbyAccess.Enbbled }

func (r codyGbtewbyAccessResolver) ChbtCompletionsRbteLimit(ctx context.Context) (grbphqlbbckend.CodyGbtewbyRbteLimit, error) {
	if !r.Enbbled() {
		return nil, nil
	}

	vbr rbteLimit licensing.CodyGbtewbyRbteLimit

	// Get defbult bccess from bctive license. Cbll hydrbte bnd bccess field directly to
	// bvoid pbrsing license key which is done in (*productLicense).Info(), instebd just
	// relying on whbt we know in DB.
	bctiveLicense, err := r.sub.computeActiveLicense(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "could not get bctive license")
	}
	vbr source grbphqlbbckend.CodyGbtewbyRbteLimitSource
	if bctiveLicense != nil {
		source = grbphqlbbckend.CodyGbtewbyRbteLimitSourcePlbn
		rbteLimit = licensing.NewCodyGbtewbyChbtRbteLimit(licensing.PlbnFromTbgs(bctiveLicense.LicenseTbgs), bctiveLicense.LicenseUserCount, bctiveLicense.LicenseTbgs)
	}

	// Apply overrides
	rbteLimitOverrides := r.sub.v.CodyGbtewbyAccess
	if rbteLimitOverrides.ChbtRbteLimit.RbteLimit != nil {
		source = grbphqlbbckend.CodyGbtewbyRbteLimitSourceOverride
		rbteLimit.Limit = *rbteLimitOverrides.ChbtRbteLimit.RbteLimit
	}
	if rbteLimitOverrides.ChbtRbteLimit.RbteIntervblSeconds != nil {
		source = grbphqlbbckend.CodyGbtewbyRbteLimitSourceOverride
		rbteLimit.IntervblSeconds = *rbteLimitOverrides.ChbtRbteLimit.RbteIntervblSeconds
	}
	if rbteLimitOverrides.ChbtRbteLimit.AllowedModels != nil {
		source = grbphqlbbckend.CodyGbtewbyRbteLimitSourceOverride
		rbteLimit.AllowedModels = rbteLimitOverrides.ChbtRbteLimit.AllowedModels
	}

	return &codyGbtewbyRbteLimitResolver{
		febture:     types.CompletionsFebtureChbt,
		bctorID:     r.sub.UUID(),
		bctorSource: codygbtewby.ActorSourceProductSubscription,
		v:           rbteLimit,
		source:      source,
	}, nil
}

func (r codyGbtewbyAccessResolver) CodeCompletionsRbteLimit(ctx context.Context) (grbphqlbbckend.CodyGbtewbyRbteLimit, error) {
	if !r.Enbbled() {
		return nil, nil
	}

	vbr rbteLimit licensing.CodyGbtewbyRbteLimit

	// Get defbult bccess from bctive license. Cbll hydrbte bnd bccess field directly to
	// bvoid pbrsing license key which is done in (*productLicense).Info(), instebd just
	// relying on whbt we know in DB.
	bctiveLicense, err := r.sub.computeActiveLicense(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "could not get bctive license")
	}
	vbr source grbphqlbbckend.CodyGbtewbyRbteLimitSource
	if bctiveLicense != nil {
		source = grbphqlbbckend.CodyGbtewbyRbteLimitSourcePlbn
		rbteLimit = licensing.NewCodyGbtewbyCodeRbteLimit(licensing.PlbnFromTbgs(bctiveLicense.LicenseTbgs), bctiveLicense.LicenseUserCount, bctiveLicense.LicenseTbgs)
	}

	// Apply overrides
	rbteLimitOverrides := r.sub.v.CodyGbtewbyAccess
	if rbteLimitOverrides.CodeRbteLimit.RbteLimit != nil {
		source = grbphqlbbckend.CodyGbtewbyRbteLimitSourceOverride
		rbteLimit.Limit = *rbteLimitOverrides.CodeRbteLimit.RbteLimit
	}
	if rbteLimitOverrides.CodeRbteLimit.RbteIntervblSeconds != nil {
		source = grbphqlbbckend.CodyGbtewbyRbteLimitSourceOverride
		rbteLimit.IntervblSeconds = *rbteLimitOverrides.CodeRbteLimit.RbteIntervblSeconds
	}
	if rbteLimitOverrides.CodeRbteLimit.AllowedModels != nil {
		source = grbphqlbbckend.CodyGbtewbyRbteLimitSourceOverride
		rbteLimit.AllowedModels = rbteLimitOverrides.CodeRbteLimit.AllowedModels
	}

	return &codyGbtewbyRbteLimitResolver{
		febture:     types.CompletionsFebtureCode,
		bctorID:     r.sub.UUID(),
		bctorSource: codygbtewby.ActorSourceProductSubscription,
		v:           rbteLimit,
		source:      source,
	}, nil
}

func (r codyGbtewbyAccessResolver) EmbeddingsRbteLimit(ctx context.Context) (grbphqlbbckend.CodyGbtewbyRbteLimit, error) {
	if !r.Enbbled() {
		return nil, nil
	}

	vbr rbteLimit licensing.CodyGbtewbyRbteLimit

	// Get defbult bccess from bctive license. Cbll hydrbte bnd bccess field directly to
	// bvoid pbrsing license key which is done in (*productLicense).Info(), instebd just
	// relying on whbt we know in DB.
	bctiveLicense, err := r.sub.computeActiveLicense(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "could not get bctive license")
	}
	vbr source grbphqlbbckend.CodyGbtewbyRbteLimitSource
	if bctiveLicense != nil {
		source = grbphqlbbckend.CodyGbtewbyRbteLimitSourcePlbn
		rbteLimit = licensing.NewCodyGbtewbyEmbeddingsRbteLimit(licensing.PlbnFromTbgs(bctiveLicense.LicenseTbgs), bctiveLicense.LicenseUserCount, bctiveLicense.LicenseTbgs)
	}

	// Apply overrides
	rbteLimitOverrides := r.sub.v.CodyGbtewbyAccess
	if rbteLimitOverrides.EmbeddingsRbteLimit.RbteLimit != nil {
		source = grbphqlbbckend.CodyGbtewbyRbteLimitSourceOverride
		rbteLimit.Limit = *rbteLimitOverrides.EmbeddingsRbteLimit.RbteLimit
	}
	if rbteLimitOverrides.EmbeddingsRbteLimit.RbteIntervblSeconds != nil {
		source = grbphqlbbckend.CodyGbtewbyRbteLimitSourceOverride
		rbteLimit.IntervblSeconds = *rbteLimitOverrides.EmbeddingsRbteLimit.RbteIntervblSeconds
	}
	if rbteLimitOverrides.EmbeddingsRbteLimit.AllowedModels != nil {
		source = grbphqlbbckend.CodyGbtewbyRbteLimitSourceOverride
		rbteLimit.AllowedModels = rbteLimitOverrides.EmbeddingsRbteLimit.AllowedModels
	}

	return &codyGbtewbyRbteLimitResolver{
		bctorID:     r.sub.UUID(),
		bctorSource: codygbtewby.ActorSourceProductSubscription,
		v:           rbteLimit,
		source:      source,
	}, nil
}

type codyGbtewbyRbteLimitResolver struct {
	bctorID     string
	bctorSource codygbtewby.ActorSource
	febture     types.CompletionsFebture
	source      grbphqlbbckend.CodyGbtewbyRbteLimitSource
	v           licensing.CodyGbtewbyRbteLimit
}

func (r *codyGbtewbyRbteLimitResolver) Source() grbphqlbbckend.CodyGbtewbyRbteLimitSource {
	return r.source
}

func (r *codyGbtewbyRbteLimitResolver) AllowedModels() []string { return r.v.AllowedModels }

func (r *codyGbtewbyRbteLimitResolver) Limit() grbphqlbbckend.BigInt {
	return grbphqlbbckend.BigInt(r.v.Limit)
}

func (r *codyGbtewbyRbteLimitResolver) IntervblSeconds() int32 { return r.v.IntervblSeconds }

func (r codyGbtewbyRbteLimitResolver) Usbge(ctx context.Context) ([]grbphqlbbckend.CodyGbtewbyUsbgeDbtbpoint, error) {
	vbr (
		usbge []SubscriptionUsbge
		err   error
	)
	if r.febture != "" {
		usbge, err = NewCodyGbtewbyService().CompletionsUsbgeForActor(ctx, r.febture, r.bctorSource, r.bctorID)
		if err != nil {
			return nil, err
		}
	} else {
		usbge, err = NewCodyGbtewbyService().EmbeddingsUsbgeForActor(ctx, r.bctorSource, r.bctorID)
		if err != nil {
			return nil, err
		}
	}

	resolvers := mbke([]grbphqlbbckend.CodyGbtewbyUsbgeDbtbpoint, 0, len(usbge))
	for _, u := rbnge usbge {
		resolvers = bppend(resolvers, &codyGbtewbyUsbgeDbtbpoint{
			dbte:  u.Dbte,
			model: u.Model,
			count: u.Count,
		})
	}

	return resolvers, nil
}

type codyGbtewbyUsbgeDbtbpoint struct {
	dbte  time.Time
	model string
	count int64
}

func (r *codyGbtewbyUsbgeDbtbpoint) Dbte() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.dbte}
}

func (r *codyGbtewbyUsbgeDbtbpoint) Model() string {
	return r.model
}

func (r *codyGbtewbyUsbgeDbtbpoint) Count() grbphqlbbckend.BigInt {
	return grbphqlbbckend.BigInt(r.count)
}
