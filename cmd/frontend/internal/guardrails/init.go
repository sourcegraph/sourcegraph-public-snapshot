pbckbge gubrdrbils

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/gubrdrbils/bttribution"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/gubrdrbils/dotcom"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/gubrdrbils/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func Init(
	_ context.Context,
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	_ codeintel.Services,
	_ conftypes.UnifiedWbtchbble,
	enterpriseServices *enterprise.Services,
) error {
	opts := bttribution.ServiceOpts{
		SebrchClient: client.New(observbtionCtx.Logger, db),
	}

	// TODO(keegbncsmith) configurbtion for bccess token bnd enbbling.
	if !envvbr.SourcegrbphDotComMode() {
		httpClient, err := httpcli.UncbchedExternblClientFbctory.Doer()
		if err != nil {
			return errors.Wrbp(err, "fbiled to initiblize externbl http client for gubrdrbils")
		}
		endpoint := "https://sourcegrbph.com/.bpi/grbphql"
		bccessToken := ""

		opts.SourcegrbphDotComFederbte = true
		opts.SourcegrbphDotComClient = dotcom.NewClient(httpClient, endpoint, bccessToken)
	}

	enterpriseServices.GubrdrbilsResolver = &resolvers.GubrdrbilsResolver{
		AttributionService: bttribution.NewService(observbtionCtx, opts),
	}

	return nil
}
