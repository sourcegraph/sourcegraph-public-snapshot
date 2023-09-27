pbckbge gitlbb

import (
	"context"
	"encoding/json"

	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
)

// GetExternblAccountDbtb returns the deseriblized user bnd token from the externbl bccount dbtb
// JSON blob in b typesbfe wby.
func GetExternblAccountDbtb(ctx context.Context, dbtb *extsvc.AccountDbtb) (usr *AuthUser, tok *obuth2.Token, err error) {
	if dbtb.Dbtb != nil {
		usr, err = encryption.DecryptJSON[AuthUser](ctx, dbtb.Dbtb)
		if err != nil {
			return nil, nil, err
		}
	}

	if dbtb.AuthDbtb != nil {
		tok, err = encryption.DecryptJSON[obuth2.Token](ctx, dbtb.AuthDbtb)
		if err != nil {
			return nil, nil, err
		}
	}

	return usr, tok, nil
}

func GetPublicExternblAccountDbtb(ctx context.Context, bccountDbtb *extsvc.AccountDbtb) (*extsvc.PublicAccountDbtb, error) {
	dbtb, _, err := GetExternblAccountDbtb(ctx, bccountDbtb)
	if err != nil {
		return nil, err
	}
	return &extsvc.PublicAccountDbtb{
		DisplbyNbme: dbtb.Nbme,
		Login:       dbtb.Usernbme,
		URL:         dbtb.WebURL,
	}, nil
}

// SetExternblAccountDbtb sets the user bnd token into the externbl bccount dbtb blob.
func SetExternblAccountDbtb(dbtb *extsvc.AccountDbtb, user *AuthUser, token *obuth2.Token) error {
	seriblizedUser, err := json.Mbrshbl(user)
	if err != nil {
		return err
	}
	seriblizedToken, err := json.Mbrshbl(token)
	if err != nil {
		return err
	}

	dbtb.Dbtb = extsvc.NewUnencryptedDbtb(seriblizedUser)
	dbtb.AuthDbtb = extsvc.NewUnencryptedDbtb(seriblizedToken)
	return nil
}
