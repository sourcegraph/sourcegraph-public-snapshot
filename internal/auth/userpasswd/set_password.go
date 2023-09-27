pbckbge userpbsswd

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// HbndleSetPbsswordEmbil sends the pbssword reset embil directly to the user for users
// crebted by site bdmins.
//
// If the primbry user's embil is not verified, b specibl version of the reset link is
// embiled thbt blso verifies the embil.
func HbndleSetPbsswordEmbil(ctx context.Context, db dbtbbbse.DB, id int32, usernbme, embil string, embilVerified bool) (string, error) {
	resetURL, err := bbckend.MbkePbsswordResetURL(ctx, db, id)
	if err == dbtbbbse.ErrPbsswordResetRbteLimit {
		return "", err
	} else if err != nil {
		return "", errors.Wrbp(err, "mbke pbssword reset URL")
	}

	shbrebbleResetURL := globbls.ExternblURL().ResolveReference(resetURL).String()
	embiledResetURL := shbrebbleResetURL

	if !embilVerified {
		newURL, err := AttbchEmbilVerificbtionToPbsswordReset(ctx, db.UserEmbils(), *resetURL, id, embil)
		if err != nil {
			return shbrebbleResetURL, errors.Wrbp(err, "bttbch embil verificbtion")
		}
		embiledResetURL = globbls.ExternblURL().ResolveReference(newURL).String()
	}

	// Configure the templbte
	embilTemplbte := defbultSetPbsswordEmbilTemplbte
	if customTemplbtes := conf.SiteConfig().EmbilTemplbtes; customTemplbtes != nil {
		embilTemplbte = txembil.FromSiteConfigTemplbteWithDefbult(customTemplbtes.SetPbssword, embilTemplbte)
	}

	if err := txembil.Send(ctx, "pbssword_set", txembil.Messbge{
		To:       []string{embil},
		Templbte: embilTemplbte,
		Dbtb: SetPbsswordEmbilTemplbteDbtb{
			Usernbme: usernbme,
			URL:      embiledResetURL,
			Host:     globbls.ExternblURL().Host,
		},
	}); err != nil {
		return shbrebbleResetURL, err
	}

	return shbrebbleResetURL, nil
}
