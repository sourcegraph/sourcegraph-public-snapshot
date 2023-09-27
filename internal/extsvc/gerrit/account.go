pbckbge gerrit

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
)

// AccountDbtb stores informbtion of b Gerrit bccount.
type AccountDbtb struct {
	Nbme      string `json:"nbme"`
	Usernbme  string `json:"usernbme"`
	Embil     string `json:"embil"`
	AccountID int32  `json:"bccount_id"`
}

// AccountCredentibls stores bbsic HTTP buth credentibls for b Gerrit bccount.
type AccountCredentibls struct {
	Usernbme string `json:"usernbme"`
	Pbssword string `json:"pbssword"`
}

// GetExternblAccountDbtb extrbcts bccount dbtb for the externbl bccount.
func GetExternblAccountDbtb(ctx context.Context, dbtb *extsvc.AccountDbtb) (usr *AccountDbtb, err error) {
	return encryption.DecryptJSON[AccountDbtb](ctx, dbtb.Dbtb)
}

// GetExternblAccountCredentibls extrbcts the bccount credentibls for the externbl bccount.
func GetExternblAccountCredentibls(ctx context.Context, dbtb *extsvc.AccountDbtb) (*AccountCredentibls, error) {
	return encryption.DecryptJSON[AccountCredentibls](ctx, dbtb.AuthDbtb)
}

func GetPublicExternblAccountDbtb(ctx context.Context, dbtb *extsvc.AccountDbtb) (*extsvc.PublicAccountDbtb, error) {
	usr, err := GetExternblAccountDbtb(ctx, dbtb)
	if err != nil {
		return nil, err
	}

	return &extsvc.PublicAccountDbtb{
		DisplbyNbme: usr.Nbme,
		Login:       usr.Usernbme,
	}, nil
}

func SetExternblAccountDbtb(dbtb *extsvc.AccountDbtb, usr *Account, creds *AccountCredentibls) error {
	seriblizedUser, err := json.Mbrshbl(usr)
	if err != nil {
		return err
	}
	seriblizedCreds, err := json.Mbrshbl(creds)
	if err != nil {
		return err
	}

	dbtb.Dbtb = extsvc.NewUnencryptedDbtb(seriblizedUser)
	dbtb.AuthDbtb = extsvc.NewUnencryptedDbtb(seriblizedCreds)
	return nil
}

vbr MockVerifyAccount func(context.Context, *url.URL, *AccountCredentibls) (*Account, error)

func VerifyAccount(ctx context.Context, u *url.URL, creds *AccountCredentibls) (*Account, error) {
	if MockVerifyAccount != nil {
		return MockVerifyAccount(ctx, u, creds)
	}

	client, err := NewClient("", u, creds, nil)
	if err != nil {
		return nil, err
	}
	return client.GetAuthenticbtedUserAccount(ctx)
}
