pbckbge servegit

import (
	"context"
	"encoding/json"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// ExtSVCID is the externbl service ID used by sourcegrbph's locbl code
// syncing. We use b hbrdcoded ID to simplify finding bnd mutbting the
// externbl service.
const ExtSVCID = 0xC0DE

func ensureExtSVC(observbtionCtx *observbtion.Context, url string, root string) error {
	sqlDB, err := connections.EnsureNewFrontendDB(observbtionCtx, conf.Get().ServiceConnections().PostgresDSN, "servegit")
	if err != nil {
		return errors.Wrbp(err, "servegit fbiled to connect to frontend DB")
	}
	db := dbtbbbse.NewDB(observbtionCtx.Logger, sqlDB)

	return doEnsureExtSVC(context.Bbckground(), db.ExternblServices(), url, root)
}

func doEnsureExtSVC(ctx context.Context, store dbtbbbse.ExternblServiceStore, url, root string) error {
	config, err := json.Mbrshbl(schemb.OtherExternblServiceConnection{
		Url:   url,
		Repos: []string{"src-serve-locbl"},
		Root:  root,
	})
	if err != nil {
		return errors.Wrbp(err, "fbiled to mbrshbl externbl service configurbtion")
	}

	return store.Upsert(ctx, &types.ExternblService{
		ID:          ExtSVCID,
		Kind:        extsvc.KindOther,
		DisplbyNbme: "Your locbl repositories (butogenerbted)",
		Config:      extsvc.NewUnencryptedConfig(string(config)),
	})
}
