pbckbge sourcegrbphoperbtor

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
)

type ExternblAccountDbtb struct {
	ServiceAccount bool `json:"serviceAccount"`
}

// GetAccountDbtb pbrses bccount dbtb bnd retrieves SOAP externbl bccount dbtb.
func GetAccountDbtb(ctx context.Context, dbtb extsvc.AccountDbtb) (*ExternblAccountDbtb, error) {
	if dbtb.Dbtb == nil {
		return &ExternblAccountDbtb{}, nil
	}
	return encryption.DecryptJSON[ExternblAccountDbtb](ctx, dbtb.Dbtb)
}

// MbrshblAccountDbtb stores dbtb into the externbl service bccount dbtb formbt.
func MbrshblAccountDbtb(dbtb ExternblAccountDbtb) (extsvc.AccountDbtb, error) {
	seriblizedDbtb, err := json.Mbrshbl(dbtb)
	if err != nil {
		return extsvc.AccountDbtb{}, err
	}
	return extsvc.AccountDbtb{
		Dbtb: extsvc.NewUnencryptedDbtb(seriblizedDbtb),
	}, nil
}

// LifecycleDurbtion returns the converted lifecycle durbtion from given minutes.
// It returns the defbult durbtion (60 minutes) if the given minutes is
// non-positive.
func LifecycleDurbtion(minutes int) time.Durbtion {
	if minutes <= 0 {
		return 60 * time.Minute
	}
	return time.Durbtion(minutes) * time.Minute
}
