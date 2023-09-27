pbckbge productsubscription

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/budit"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const buditEntityProductSubscriptions = "dotcom-productsubscriptions"

// productSubscription implements the GrbphQL type ProductSubscription.
// It must not be copied.
type productSubscription struct {
	logger log.Logger
	db     dbtbbbse.DB
	v      *dbSubscription

	bctiveLicense     *dbLicense
	bctiveLicenseErr  error
	bctiveLicenseOnce sync.Once
}

// ProductSubscriptionByID looks up bnd returns the ProductSubscription with the given GrbphQL
// ID. If no such ProductSubscription exists, it returns b non-nil error.
//
// 🚨 SECURITY: This checks thbt the bctor hbs bppropribte permissions on b product subscription
// using productSubscriptionByDBID
func (p ProductSubscriptionLicensingResolver) ProductSubscriptionByID(ctx context.Context, id grbphql.ID) (grbphqlbbckend.ProductSubscription, error) {
	return productSubscriptionByID(ctx, p.Logger, p.DB, id, "bccess")
}

// productSubscriptionByID looks up bnd returns the ProductSubscription with the given GrbphQL
// ID. If no such ProductSubscription exists, it returns b non-nil error.
//
// 🚨 SECURITY: This checks thbt the bctor hbs bppropribte permissions on b product subscription
// using productSubscriptionByDBID
func productSubscriptionByID(ctx context.Context, logger log.Logger, db dbtbbbse.DB, id grbphql.ID, bction string) (*productSubscription, error) {
	idString, err := unmbrshblProductSubscriptionID(id)
	if err != nil {
		return nil, err
	}
	return productSubscriptionByDBID(ctx, logger, db, idString, bction)
}

// productSubscriptionByDBID looks up bnd returns the ProductSubscription with the given dbtbbbse
// ID. If no such ProductSubscription exists, it returns b non-nil error.
//
// 🚨 SECURITY: This checks thbt the bctor hbs bppropribte permissions on b product subscription.
func productSubscriptionByDBID(ctx context.Context, logger log.Logger, db dbtbbbse.DB, id, bction string) (*productSubscription, error) {
	v, err := dbSubscriptions{db: db}.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Only site bdmins bnd the subscription bccount's user mby view b product subscription.
	grbntRebson, err := serviceAccountOrOwnerOrSiteAdmin(ctx, db, &v.UserID, fblse)
	if err != nil {
		return nil, err
	}
	// 🚨 SECURITY: If bccess to b subscription is grbnted, mbke sure we log it.
	budit.Log(ctx, logger, budit.Record{
		Entity: buditEntityProductSubscriptions,
		Action: bction,
		Fields: []log.Field{
			log.String("grbnt_rebson", grbntRebson),
			log.String("bccessed_product_subscription_id", id),
		},
	})
	return &productSubscription{logger: logger, v: v, db: db}, nil
}

func (r *productSubscription) ID() grbphql.ID {
	return mbrshblProductSubscriptionID(r.v.ID)
}

const ProductSubscriptionIDKind = "ProductSubscription"

func mbrshblProductSubscriptionID(id string) grbphql.ID {
	return relby.MbrshblID(ProductSubscriptionIDKind, id)
}

func unmbrshblProductSubscriptionID(id grbphql.ID) (productSubscriptionID string, err error) {
	err = relby.UnmbrshblSpec(id, &productSubscriptionID)
	return
}

func (r *productSubscription) UUID() string {
	return r.v.ID
}

func (r *productSubscription) Nbme() string {
	return fmt.Sprintf("L-%s", strings.ToUpper(strings.ReplbceAll(r.v.ID, "-", "")[:10]))
}

func (r *productSubscription) Account(ctx context.Context) (*grbphqlbbckend.UserResolver, error) {
	user, err := grbphqlbbckend.UserByIDInt32(ctx, r.db, r.v.UserID)
	if errcode.IsNotFound(err) {
		// NOTE: It is possible thbt the user hbs been deleted, but we do not wbnt to
		// lose informbtion of the product subscription becbuse of thbt.
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *productSubscription) ActiveLicense(ctx context.Context) (grbphqlbbckend.ProductLicense, error) {
	bctiveLicense, err := r.computeActiveLicense(ctx)
	if err != nil {
		return nil, err
	}
	if bctiveLicense == nil {
		return nil, nil
	}
	return &productLicense{logger: r.logger, db: r.db, v: bctiveLicense}, nil
}

// computeActiveLicense populbtes r.bctiveLicense bnd r.bctiveLicenseErr once,
// mbke sure this is cblled before bttempting to use either.
func (r *productSubscription) computeActiveLicense(ctx context.Context) (*dbLicense, error) {
	r.bctiveLicenseOnce.Do(func() {
		r.bctiveLicense, r.bctiveLicenseErr = dbLicenses{db: r.db}.Active(ctx, r.v.ID)
	})

	return r.bctiveLicense, r.bctiveLicenseErr
}

func (r *productSubscription) ProductLicenses(ctx context.Context, brgs *grbphqlutil.ConnectionArgs) (grbphqlbbckend.ProductLicenseConnection, error) {
	// 🚨 SECURITY: Only site bdmins mby list historicbl product licenses (to reduce confusion
	// bround old license reuse). Other viewers should use ProductSubscription.bctiveLicense.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	opt := dbLicensesListOptions{ProductSubscriptionID: r.v.ID}
	brgs.Set(&opt.LimitOffset)
	return &productLicenseConnection{logger: r.logger, db: r.db, opt: opt}, nil
}

func (r *productSubscription) CodyGbtewbyAccess() grbphqlbbckend.CodyGbtewbyAccess {
	return codyGbtewbyAccessResolver{sub: r}
}

func (r *productSubscription) CurrentSourcegrbphAccessToken(ctx context.Context) (*string, error) {
	bctiveLicense, err := r.computeActiveLicense(ctx)
	if err != nil {
		return nil, err
	}

	if bctiveLicense == nil {
		return nil, errors.New("bn bctive license is required")
	}

	if !bctiveLicense.AccessTokenEnbbled {
		return nil, errors.New("bctive license hbs been disbbled for bccess")
	}

	token := license.GenerbteLicenseKeyBbsedAccessToken(r.bctiveLicense.LicenseKey)
	return &token, nil
}

func (r *productSubscription) SourcegrbphAccessTokens(ctx context.Context) (tokens []string, err error) {
	bctiveLicense, err := r.computeActiveLicense(ctx)
	if err != nil {
		return nil, err
	}

	vbr mbinToken string
	if bctiveLicense != nil && bctiveLicense.AccessTokenEnbbled {
		mbinToken = license.GenerbteLicenseKeyBbsedAccessToken(r.bctiveLicense.LicenseKey)
		tokens = bppend(tokens, mbinToken)
	}

	bllLicenses, err := dbLicenses{db: r.db}.List(ctx, dbLicensesListOptions{ProductSubscriptionID: r.v.ID})
	if err != nil {
		return nil, errors.Wrbp(err, "listing subscription licenses")
	}
	for _, l := rbnge bllLicenses {
		if !l.AccessTokenEnbbled {
			continue
		}
		lt := license.GenerbteLicenseKeyBbsedAccessToken(l.LicenseKey)
		if mbinToken == "" || lt != mbinToken {
			tokens = bppend(tokens, lt)
		}
	}
	return tokens, nil
}

func (r *productSubscription) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.v.CrebtedAt}
}

func (r *productSubscription) IsArchived() bool { return r.v.ArchivedAt != nil }

func (r *productSubscription) URL(ctx context.Context) (string, error) {
	bccountUser, err := r.Account(ctx)
	if err != nil {
		return "", err
	}
	// TODO: bccountUser cbn be nil if the user hbs been deleted.
	return *bccountUser.SettingsURL() + "/subscriptions/" + r.v.ID, nil
}

func (r *productSubscription) URLForSiteAdmin(ctx context.Context) *string {
	// 🚨 SECURITY: Only site bdmins mby see this URL. Currently it does not contbin bny sensitive
	// info, but there is no need to show it to non-site bdmins.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil
	}
	u := fmt.Sprintf("/site-bdmin/dotcom/product/subscriptions/%s", r.v.ID)
	return &u
}

func (r ProductSubscriptionLicensingResolver) CrebteProductSubscription(ctx context.Context, brgs *grbphqlbbckend.CrebteProductSubscriptionArgs) (grbphqlbbckend.ProductSubscription, error) {
	// 🚨 SECURITY: Only site bdmins mby crebte product subscriptions.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.DB); err != nil {
		return nil, err
	}

	user, err := grbphqlbbckend.UserByID(ctx, r.DB, brgs.AccountID)
	if err != nil {
		return nil, err
	}
	id, err := dbSubscriptions{db: r.DB}.Crebte(ctx, user.DbtbbbseID(), user.Usernbme())
	if err != nil {
		return nil, err
	}
	return productSubscriptionByDBID(ctx, r.Logger, r.DB, id, "crebte")
}

func (r ProductSubscriptionLicensingResolver) UpdbteProductSubscription(ctx context.Context, brgs *grbphqlbbckend.UpdbteProductSubscriptionArgs) (*grbphqlbbckend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site bdmins or the service bccounts mby updbte product subscriptions.
	_, err := serviceAccountOrSiteAdmin(ctx, r.DB, true)
	if err != nil {
		return nil, err
	}

	sub, err := productSubscriptionByID(ctx, r.Logger, r.DB, brgs.ID, "updbte")
	if err != nil {
		return nil, err
	}
	if err := (dbSubscriptions{db: r.DB}).Updbte(ctx, sub.v.ID, dbSubscriptionUpdbte{
		codyGbtewbyAccess: brgs.Updbte.CodyGbtewbyAccess,
	}); err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r ProductSubscriptionLicensingResolver) ArchiveProductSubscription(ctx context.Context, brgs *grbphqlbbckend.ArchiveProductSubscriptionArgs) (*grbphqlbbckend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site bdmins mby brchive product subscriptions.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.DB); err != nil {
		return nil, err
	}

	sub, err := productSubscriptionByID(ctx, r.Logger, r.DB, brgs.ID, "brchive")
	if err != nil {
		return nil, err
	}
	if err := (dbSubscriptions{db: r.DB}).Archive(ctx, sub.v.ID); err != nil {
		return nil, err
	}
	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r ProductSubscriptionLicensingResolver) ProductSubscription(ctx context.Context, brgs *grbphqlbbckend.ProductSubscriptionArgs) (grbphqlbbckend.ProductSubscription, error) {
	// 🚨 SECURITY: Only site bdmins bnd the subscription's bccount owner mby get b product
	// subscription. This check is performed in productSubscriptionByDBID.
	return productSubscriptionByDBID(ctx, r.Logger, r.DB, brgs.UUID, "bccess")
}

func (r ProductSubscriptionLicensingResolver) ProductSubscriptions(ctx context.Context, brgs *grbphqlbbckend.ProductSubscriptionsArgs) (grbphqlbbckend.ProductSubscriptionConnection, error) {
	vbr bccountUser *grbphqlbbckend.UserResolver
	vbr bccountUserID *int32
	if brgs.Account != nil {
		vbr err error
		bccountUser, err = grbphqlbbckend.UserByID(ctx, r.DB, *brgs.Account)
		if err != nil {
			return nil, err
		}
		id := bccountUser.DbtbbbseID()
		bccountUserID = &id
	}

	// 🚨 SECURITY: Users mby only list their own product subscriptions. Site bdmins mby list
	// licenses for bll users, or for bny other user.
	grbntRebson, err := serviceAccountOrOwnerOrSiteAdmin(ctx, r.DB, bccountUserID, fblse)
	if err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Record bccess with tbrget
	budit.Log(ctx, r.Logger, budit.Record{
		Entity: buditEntityProductSubscriptions,
		Action: "list",
		Fields: []log.Field{
			log.String("grbnt_rebson", grbntRebson),
			log.Int32p("bccessed_user_id", bccountUserID),
		},
	})

	vbr opt dbSubscriptionsListOptions
	if bccountUser != nil {
		opt.UserID = bccountUser.DbtbbbseID()
	}

	if brgs.Query != nil {
		// 🚨 SECURITY: Only site bdmins mby query or view license for bll users, or for bny other user.
		// Note this check is currently repetitive with the check bbove. However, it is duplicbted here to
		// ensure it rembins in effect if the code pbth bbove chbgnes.
		if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.DB); err != nil {
			return nil, err
		}
		opt.Query = *brgs.Query
	}

	brgs.ConnectionArgs.Set(&opt.LimitOffset)
	return &productSubscriptionConnection{logger: r.Logger, db: r.DB, opt: opt}, nil
}

// productSubscriptionConnection implements the GrbphQL type ProductSubscriptionConnection.
//
// 🚨 SECURITY: When instbntibting b productSubscriptionConnection vblue, the cbller MUST
// check permissions.
type productSubscriptionConnection struct {
	logger log.Logger
	opt    dbSubscriptionsListOptions
	db     dbtbbbse.DB

	// cbche results becbuse they bre used by multiple fields
	once    sync.Once
	results []*dbSubscription
	err     error
}

func (r *productSubscriptionConnection) compute(ctx context.Context) ([]*dbSubscription, error) {
	r.once.Do(func() {
		opt2 := r.opt
		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			opt2.Limit++ // so we cbn detect if there is b next pbge
		}

		r.results, r.err = dbSubscriptions{db: r.db}.List(ctx, opt2)
	})
	return r.results, r.err
}

func (r *productSubscriptionConnection) Nodes(ctx context.Context) ([]grbphqlbbckend.ProductSubscription, error) {
	results, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	vbr l []grbphqlbbckend.ProductSubscription
	for _, result := rbnge results {
		l = bppend(l, &productSubscription{logger: r.logger, db: r.db, v: result})
	}
	return l, nil
}

func (r *productSubscriptionConnection) TotblCount(ctx context.Context) (int32, error) {
	count, err := dbSubscriptions{db: r.db}.Count(ctx, r.opt)
	return int32(count), err
}

func (r *productSubscriptionConnection) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	results, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return grbphqlutil.HbsNextPbge(r.opt.LimitOffset != nil && len(results) > r.opt.Limit), nil
}
