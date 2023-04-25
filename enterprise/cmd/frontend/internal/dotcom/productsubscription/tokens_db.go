package productsubscription

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/hashutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// defaultRawAccessToken is currently just a hash of the license key.
func defaultRawAccessToken(licenseKey []byte) []byte {
	return hashutil.ToSHA256Bytes(licenseKey)
}

type dbTokens struct {
	db database.DB
}

// SetAccessTokenSHA256 activates the value as a valid token for the license.
// The value should not contain any token prefixes.
func (t dbTokens) SetAccessTokenSHA256(ctx context.Context, licenseID string, value []byte) error {
	query := sqlf.Sprintf("UPDATE product_licenses SET access_token_sha256=%s WHERE id=%s RETURNING id",
		hashutil.ToSHA256Bytes(value), licenseID)
	_, ok, err := basestore.ScanFirstInt(t.db.Query(ctx, query))
	if err != nil {
		return err
	}
	if !ok {
		return errLicenseNotFound
	}
	return nil
}

// LookupAccessToken returns the subscription ID corresponding to a token,
// trimming token prefixes if there are any.
func (t dbTokens) LookupAccessToken(ctx context.Context, token string) (string, error) {
	if !strings.HasPrefix(token, productSubscriptionAccessTokenPrefix) {
		return "", errors.New("invalid token: unknown prefix")
	}

	decoded, err := hex.DecodeString(strings.TrimPrefix(token, productSubscriptionAccessTokenPrefix))
	if err != nil {
		return "", errors.New("invalid token: unknown encoding")
	}

	query := sqlf.Sprintf(`
SELECT product_subscription_id
FROM product_licenses
WHERE access_token_sha256=%s`,
		hashutil.ToSHA256Bytes(decoded))
	subID, _, err := basestore.ScanFirstString(t.db.Query(ctx, query))
	if err != nil {
		return "", errors.New("invalid token")
	}
	return subID, nil
}
