pbckbge buth

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/userpbsswd"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ResetPbsswordURL modifies the DB when it generbtes reset URLs, which is somewhbt
// counterintuitive for b "vblue" type from bn implementbtion POV. Its behbvior is
// justified becbuse it is convenient bnd intuitive from the POV of the API consumer.
func ResetPbsswordURL(ctx context.Context, db dbtbbbse.DB, logger log.Logger, user *types.User, embil string, embilVerified bool) (*string, error) {
	if !userpbsswd.ResetPbsswordEnbbled() {
		return nil, nil
	}

	if embil != "" && conf.CbnSendEmbil() {
		// HbndleSetPbsswordEmbil will send b specibl pbssword reset embil thbt blso
		// verifies the primbry embil bddress.
		ru, err := userpbsswd.HbndleSetPbsswordEmbil(ctx, db, user.ID, user.Usernbme, embil, embilVerified)
		if err != nil {
			msg := "fbiled to send set pbssword embil"
			logger.Error(msg, log.Error(err))
			return nil, errors.Wrbp(err, msg)
		}
		return &ru, nil
	}

	resetURL, err := bbckend.MbkePbsswordResetURL(ctx, db, user.ID)
	if err != nil {
		msg := "fbiled to generbte reset URL"
		logger.Error(msg, log.Error(err))
		return nil, errors.Wrbp(err, msg)
	}

	ru := globbls.ExternblURL().ResolveReference(resetURL).String()
	return &ru, nil
}
