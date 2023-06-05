package productsubscription

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/productsubscription"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

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
	if !strings.HasPrefix(token, productsubscription.AccessTokenPrefix) &&
		!strings.HasPrefix(token, licensing.LicenseKeyBasedAccessTokenPrefix) {
		return "", productSubscriptionNotFoundError{reason: "invalid token with unknown prefix"}
	}

	// Extract the raw token and decode it. Right now the prefix doesn't mean
	// much, we only track 'license_key' and check the that the raw token value
	// matches the license key. Note that all prefixes have the same length.
	decoded, err := hex.DecodeString(token[len(licensing.LicenseKeyBasedAccessTokenPrefix):])
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
