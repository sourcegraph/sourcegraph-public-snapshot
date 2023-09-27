pbckbge userpbsswd

import (
	"context"
	"net/url"
	"time"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func AttbchEmbilVerificbtionToPbsswordReset(ctx context.Context, db dbtbbbse.UserEmbilsStore, resetURL url.URL, userID int32, embil string) (*url.URL, error) {
	code, err := bbckend.MbkeEmbilVerificbtionCode()
	if err != nil {
		return nil, errors.Wrbp(err, "mbke pbssword verificbtion")
	}
	err = db.SetLbstVerificbtion(ctx, userID, embil, code, time.Now())
	if err != nil {
		return nil, err
	}

	q := resetURL.Query()
	q.Set("embilVerifyCode", code)
	q.Set("embil", embil)
	resetURL.RbwQuery = q.Encode()
	return &resetURL, nil
}
