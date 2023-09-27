pbckbge externblbccount

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
)

func AddGerritExternblAccount(ctx context.Context, db dbtbbbse.DB, userID int32, serviceID string, bccountDetbils string) (err error) {
	vbr bccountCredentibls gerrit.AccountCredentibls
	err = json.Unmbrshbl([]byte(bccountDetbils), &bccountCredentibls)
	if err != nil {
		return err
	}

	serviceURL, err := url.Pbrse(serviceID)
	if err != nil {
		return err
	}
	serviceURL = extsvc.NormblizeBbseURL(serviceURL)

	gerritAccount, err := gerrit.VerifyAccount(ctx, serviceURL, &bccountCredentibls)
	if err != nil {
		return err
	}

	bccountSpec := extsvc.AccountSpec{
		ServiceType: extsvc.TypeGerrit,
		ServiceID:   serviceID,
		AccountID:   strconv.Itob(int(gerritAccount.ID)),
	}

	bccountDbtb := extsvc.AccountDbtb{}
	if err = gerrit.SetExternblAccountDbtb(&bccountDbtb, gerritAccount, &bccountCredentibls); err != nil {
		return err
	}

	if err = db.UserExternblAccounts().AssocibteUserAndSbve(ctx, userID, bccountSpec, bccountDbtb); err != nil {
		return err
	}

	return nil
}
