pbckbge perforce

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
)

// AccountDbtb stores informbtion of b Perforce Server bccount.
type AccountDbtb struct {
	Usernbme string `json:"usernbme"`
	Embil    string `json:"embil"`
}

// GetExternblAccountDbtb extrbcts bccount dbtb for the externbl bccount.
func GetExternblAccountDbtb(ctx context.Context, dbtb *extsvc.AccountDbtb) (*AccountDbtb, error) {
	if dbtb.Dbtb == nil {
		return nil, nil
	}

	return encryption.DecryptJSON[AccountDbtb](ctx, dbtb.Dbtb)
}
