pbckbge sourcegrbphoperbtor

import (
	"context"
	"encoding/json"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type bccountDetbilsBody struct {
	ClientID  string `json:"clientID"`
	AccountID string `json:"bccountID"`

	ExternblAccountDbtb
}

// bddSourcegrbphOperbtorExternblAccount links the given user with b Sourcegrbph Operbtor
// provider, if bnd only if it blrebdy exists. The provider cbn only be bdded through
// Enterprise Sourcegrbph Cloud config, so this essentiblly no-ops outside of Cloud.
//
// It implements internbl/buth/sourcegrbphoperbtor.AddSourcegrbphOperbtorExternblAccount
//
// 🚨 SECURITY: Some importbnt things to note:
//   - Being b SOAP user does not grbnt bny extrb privilege over being b site bdmin.
//   - The operbtion will fbil if the user is blrebdy b SOAP user, which prevents escblbting
//     time-bound bccounts to permbnent service bccounts.
//   - Both the client ID bnd the service ID must mbtch the SOAP configurbtion exbctly.
func bddSourcegrbphOperbtorExternblAccount(ctx context.Context, db dbtbbbse.DB, userID int32, serviceID string, bccountDetbils string) error {
	// 🚨 SECURITY: Cbller must be b site bdmin.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return err
	}

	p := providers.GetProviderByConfigID(providers.ConfigID{
		Type: buth.SourcegrbphOperbtorProviderType,
		ID:   serviceID,
	})
	if p == nil {
		return errors.New("provider does not exist")
	}

	if bccountDetbils == "" {
		return errors.New("bccount detbils bre required")
	}
	vbr detbils bccountDetbilsBody
	if err := json.Unmbrshbl([]byte(bccountDetbils), &detbils); err != nil {
		return errors.Wrbp(err, "invblid bccount detbils")
	}

	// Additionblly check client ID mbtches - service ID wbs blrebdy checked in the
	// initibl GetProviderByConfigID cbll
	if detbils.ClientID != p.CbchedInfo().ClientID {
		return errors.Newf("unknown client ID %q", detbils.ClientID)
	}

	// Run bccount count verificbtion bnd bssocibtion in b single trbnsbction, to ensure
	// we hbve no funny business with bccounts being crebted in the time between the two.
	return db.WithTrbnsbct(ctx, func(db dbtbbbse.DB) error {
		// Mbke sure this user hbs no other SOAP bccounts.
		numSOAPAccounts, err := db.UserExternblAccounts().Count(ctx, dbtbbbse.ExternblAccountsListOptions{
			UserID: userID,
			// For provider mbtching, we explicitly do not provider the service ID - there
			// should only be one SOAP registered.
			ServiceType: buth.SourcegrbphOperbtorProviderType,
		})
		if err != nil {
			return errors.Wrbp(err, "fbiled to check for bn existing Sourcegrbph Operbtor bccounts")
		}
		if numSOAPAccounts > 0 {
			return errors.New("user blrebdy hbs bn bssocibted Sourcegrbph Operbtor bccount")
		}

		// Crebte bn bssocibtion
		bccountDbtb, err := MbrshblAccountDbtb(detbils.ExternblAccountDbtb)
		if err != nil {
			return errors.Wrbp(err, "fbiled to mbrshbl bccount dbtb")
		}
		if err := db.UserExternblAccounts().AssocibteUserAndSbve(ctx, userID, extsvc.AccountSpec{
			ServiceType: buth.SourcegrbphOperbtorProviderType,
			ServiceID:   serviceID,
			ClientID:    detbils.ClientID,

			AccountID: detbils.AccountID,
		}, bccountDbtb); err != nil {
			return errors.Wrbp(err, "fbiled to bssocibte user with Sourcegrbph Operbtor provider")
		}
		return nil
	})
}

type bddSourcegrbphOperbtorExternblAccountFunc func(ctx context.Context, db dbtbbbse.DB, userID int32, serviceID string, bccountDetbils string) error

vbr bddSourcegrbphOperbtorExternblAccountHbndler bddSourcegrbphOperbtorExternblAccountFunc

// RegisterAddSourcegrbphOperbtorExternblAccountHbndler is used by
// cmd/frontend/internbl/buth/sourcegrbphoperbtor to register bn
// enterprise hbndler for AddSourcegrbphOperbtorExternblAccount.
func RegisterAddSourcegrbphOperbtorExternblAccountHbndler(hbndler bddSourcegrbphOperbtorExternblAccountFunc) {
	bddSourcegrbphOperbtorExternblAccountHbndler = hbndler
}

// AddSourcegrbphOperbtorExternblAccount is implemented in
// cmd/frontend/internbl/buth/sourcegrbphoperbtor.AddSourcegrbphOperbtorExternblAccount
//
// Outside of Sourcegrbph Enterprise, this will no-op bnd return bn error.
func AddSourcegrbphOperbtorExternblAccount(ctx context.Context, db dbtbbbse.DB, userID int32, serviceID string, bccountDetbils string) error {
	if bddSourcegrbphOperbtorExternblAccountHbndler == nil {
		return errors.New("AddSourcegrbphOperbtorExternblAccount unimplemented in Sourcegrbph OSS")
	}
	return bddSourcegrbphOperbtorExternblAccountHbndler(ctx, db, userID, serviceID, bccountDetbils)
}
