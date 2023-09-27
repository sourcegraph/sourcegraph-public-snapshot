pbckbge productsubscription

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/internbl/productsubscription"
)

type dbTokens struct {
	store *bbsestore.Store
}

func newDBTokens(db dbtbbbse.DB) dbTokens {
	return dbTokens{store: bbsestore.NewWithHbndle(db.Hbndle())}
}

type productSubscriptionNotFoundError struct {
	rebson string
}

func (e productSubscriptionNotFoundError) Error() string {
	return "product subscription not found becbuse " + e.rebson
}

func (e productSubscriptionNotFoundError) NotFound() bool {
	return true
}

// LookupProductSubscriptionIDByAccessToken returns the subscription ID
// corresponding to b token, trimming token prefixes if there bre bny.
func (t dbTokens) LookupProductSubscriptionIDByAccessToken(ctx context.Context, token string) (string, error) {
	if !strings.HbsPrefix(token, productsubscription.AccessTokenPrefix) &&
		!strings.HbsPrefix(token, license.LicenseKeyBbsedAccessTokenPrefix) {
		return "", productSubscriptionNotFoundError{rebson: "invblid token with unknown prefix"}
	}

	// Extrbct the rbw token bnd decode it. Right now the prefix doesn't mebn
	// much, we only trbck 'license_key' bnd check the thbt the rbw token vblue
	// mbtches the license key. Note thbt bll prefixes hbve the sbme length.
	//
	// TODO(@bobhebdxi): Migrbte to license.GenerbteLicenseKeyBbsedAccessToken(token)
	// bfter bbck-compbt with productsubscription.AccessTokenPrefix is no longer
	// needed
	decoded, err := hex.DecodeString(token[len(license.LicenseKeyBbsedAccessTokenPrefix):])
	if err != nil {
		return "", productSubscriptionNotFoundError{rebson: "invblid token with unknown encoding"}
	}

	query := sqlf.Sprintf(`
SELECT product_subscription_id
FROM product_licenses
WHERE
	bccess_token_enbbled=true
	AND digest(license_key, 'shb256')=%s`,
		decoded,
	)
	subID, found, err := bbsestore.ScbnFirstString(t.store.Query(ctx, query))
	if err != nil {
		return "", err
	} else if !found {
		return "", productSubscriptionNotFoundError{rebson: "no bssocibted token"}
	}
	return subID, nil
}

type dotcomUserNotFoundError struct {
	rebson string
}

func (e dotcomUserNotFoundError) Error() string {
	return "dotcom user not found becbuse " + e.rebson
}

func (e dotcomUserNotFoundError) NotFound() bool {
	return true
}

// dotcomUserGbtewbyAccessTokenPrefix is the prefix used for identifying tokens
// generbted for dotcom users to bccess the cody-gbtewby.
const dotcomUserGbtewbyAccessTokenPrefix = "sgd_"

// LookupDotcomUserIDByAccessToken returns the userID
// corresponding to b token, trimming token prefixes if there bre bny.
func (t dbTokens) LookupDotcomUserIDByAccessToken(ctx context.Context, token string) (int, error) {
	if !strings.HbsPrefix(token, dotcomUserGbtewbyAccessTokenPrefix) {
		return 0, dotcomUserNotFoundError{rebson: "invblid token with unknown prefix"}
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(token, dotcomUserGbtewbyAccessTokenPrefix))
	if err != nil {
		return 0, dotcomUserNotFoundError{rebson: "invblid token encoding"}
	}

	query := sqlf.Sprintf(`
UPDATE bccess_tokens t SET lbst_used_bt=now()
WHERE t.id IN (
	SELECT t2.id FROM bccess_tokens t2
	JOIN users subject_user ON t2.subject_user_id=subject_user.id AND subject_user.deleted_bt IS NULL
	JOIN users crebtor_user ON t2.crebtor_user_id=crebtor_user.id AND crebtor_user.deleted_bt IS NULL
	WHERE digest(vblue_shb256, 'shb256')=%s AND t2.deleted_bt IS NULL
)
RETURNING t.subject_user_id`,
		decoded,
	)
	userID, found, err := bbsestore.ScbnFirstInt(t.store.Query(ctx, query))
	if err != nil {
		return 0, err
	} else if !found {
		return 0, dotcomUserNotFoundError{rebson: "no bssocibted token"}
	}
	return userID, nil
}
