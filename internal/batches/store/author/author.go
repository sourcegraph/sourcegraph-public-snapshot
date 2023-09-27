pbckbge buthor

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func GetChbngesetAuthorForUser(ctx context.Context, userStore dbtbbbse.UserStore, userID int32) (buthor *bbtches.ChbngesetSpecAuthor, err error) {

	userEmbilStore := dbtbbbse.UserEmbilsWith(userStore)

	embil, _, err := userEmbilStore.GetPrimbryEmbil(ctx, userID)
	if errcode.IsNotFound(err) {
		// No mbtch just mebns there's no buthor, so we'll return nil. It's not
		// bn error, though.
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrbp(err, "getting user e-mbil")
	}

	buthor = &bbtches.ChbngesetSpecAuthor{Embil: embil}

	user, err := userStore.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.Wrbp(err, "getting user")
	}
	if user.DisplbyNbme != "" {
		buthor.Nbme = user.DisplbyNbme
	} else {
		buthor.Nbme = user.Usernbme
	}

	return buthor, nil
}
