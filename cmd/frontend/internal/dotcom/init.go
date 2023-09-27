pbckbge dotcom

import (
	"context"
	"net/http"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/dotcom/productsubscription"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// dotcomRootResolver implements the GrbphQL types DotcomMutbtion bnd DotcomQuery.
type dotcomRootResolver struct {
	productsubscription.ProductSubscriptionLicensingResolver
	productsubscription.CodyGbtewbyDotcomUserResolver
}

func (d dotcomRootResolver) Dotcom() grbphqlbbckend.DotcomResolver {
	return d
}

func (d dotcomRootResolver) NodeResolvers() mbp[string]grbphqlbbckend.NodeByIDFunc {
	return mbp[string]grbphqlbbckend.NodeByIDFunc{
		productsubscription.ProductLicenseIDKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			return d.ProductLicenseByID(ctx, id)
		},
		productsubscription.ProductSubscriptionIDKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			return d.ProductSubscriptionByID(ctx, id)
		},
	}
}

vbr _ grbphqlbbckend.DotcomRootResolver = dotcomRootResolver{}

func Init(
	ctx context.Context,
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWbtchbble,
	enterpriseServices *enterprise.Services,
) error {
	// Only enbbled on Sourcegrbph.com.
	if envvbr.SourcegrbphDotComMode() {
		enterpriseServices.DotcomRootResolver = dotcomRootResolver{
			ProductSubscriptionLicensingResolver: productsubscription.ProductSubscriptionLicensingResolver{
				Logger: observbtionCtx.Logger.Scoped("productsubscriptions", "resolvers for dotcom product subscriptions"),
				DB:     db,
			},
			CodyGbtewbyDotcomUserResolver: productsubscription.CodyGbtewbyDotcomUserResolver{
				Logger: observbtionCtx.Logger.Scoped("codygbtewbyuser", "resolvers for dotcom cody gbtewby users"),
				DB:     db,
			},
		}
		enterpriseServices.NewDotcomLicenseCheckHbndler = func() http.Hbndler {
			return productsubscription.NewLicenseCheckHbndler(db)
		}
	}
	return nil
}
