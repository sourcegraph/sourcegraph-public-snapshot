package productsubscription

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/hashutil"
)

// defaultRawAccessToken is currently just a hash of the license key.
func defaultRawAccessToken(licenseKey []byte) []byte {
	return hashutil.ToSHA256Bytes(licenseKey)
}

type dbTokens struct {
	store *basestore.Store
}

func newDBTokens(db database.DB) dbTokens {
	return dbTokens{store: basestore.NewWithHandle(db.Handle())}
}

type productSubscriptionNotFoundError struct {
	reason string
}

func (e productSubscriptionNotFoundError) Error() string {
	return "product subscription not found because " + e.reason
}

func (e productSubscriptionNotFoundError) NotFound() bool {
	return true
}

// LookupProductSubscriptionIDByAccessToken returns the subscription ID
// corresponding to a token, trimming token prefixes if there are any.
func (t dbTokens) LookupProductSubscriptionIDByAccessToken(ctx context.Context, token string) (string, error) {
	if !strings.HasPrefix(token, productSubscriptionAccessTokenPrefix) {
		return "", productSubscriptionNotFoundError{reason: "invalid token with unknown prefix"}
	}

	decoded, err := hex.DecodeString(strings.TrimPrefix(token, productSubscriptionAccessTokenPrefix))
	if err != nil {
		return "", productSubscriptionNotFoundError{reason: "invalid token with unknown encoding"}
	}

	query := sqlf.Sprintf(`
SELECT product_subscription_id
FROM product_licenses
WHERE
	access_token_enabled=true
	AND digest(license_key, 'sha256')=%s`,
		decoded,
	)
	subID, found, err := basestore.ScanFirstString(t.store.Query(ctx, query))
	if err != nil {
		return "", err
	} else if !found {
		return "", productSubscriptionNotFoundError{reason: "no associated token"}
	}
	return subID, nil
}
