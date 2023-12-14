package productsubscription

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/accesstoken"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/productsubscription"
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
		!strings.HasPrefix(token, license.LicenseKeyBasedAccessTokenPrefix) {
		return "", productSubscriptionNotFoundError{reason: "invalid token with unknown prefix"}
	}

	// Extract the raw token and decode it. Right now the prefix doesn't mean
	// much, we only track 'license_key' and check the that the raw token value
	// matches the license key. Note that all prefixes have the same length.
	//
	// TODO(@bobheadxi): Migrate to license.GenerateLicenseKeyBasedAccessToken(token)
	// after back-compat with productsubscription.AccessTokenPrefix is no longer
	// needed
	decoded, err := hex.DecodeString(token[len(license.LicenseKeyBasedAccessTokenPrefix):])
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

type dotcomUserNotFoundError struct {
	reason string
}

func (e dotcomUserNotFoundError) Error() string {
	return "dotcom user not found because " + e.reason
}

func (e dotcomUserNotFoundError) NotFound() bool {
	return true
}

// LookupDotcomUserIDByAccessToken returns the userID
// corresponding to a token, trimming token prefixes if there are any.
func (t dbTokens) LookupDotcomUserIDByAccessToken(ctx context.Context, token string) (int, error) {
	if !strings.HasPrefix(token, accesstoken.DotcomUserGatewayAccessTokenPrefix) {
		return 0, dotcomUserNotFoundError{reason: "invalid token with unknown prefix"}
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(token, accesstoken.DotcomUserGatewayAccessTokenPrefix))
	if err != nil {
		return 0, dotcomUserNotFoundError{reason: "invalid token encoding"}
	}

	query := sqlf.Sprintf(`
UPDATE access_tokens t SET last_used_at=now()
WHERE t.id IN (
	SELECT t2.id FROM access_tokens t2
	JOIN users subject_user ON t2.subject_user_id=subject_user.id AND subject_user.deleted_at IS NULL
	JOIN users creator_user ON t2.creator_user_id=creator_user.id AND creator_user.deleted_at IS NULL
	WHERE
		digest(value_sha256, 'sha256')=%s
		AND
		t2.deleted_at IS NULL
		AND
		(t2.expires_at IS NULL OR t2.expires_at > NOW())
)
RETURNING t.subject_user_id`,
		decoded,
	)
	userID, found, err := basestore.ScanFirstInt(t.store.Query(ctx, query))
	if err != nil {
		return 0, err
	} else if !found {
		return 0, dotcomUserNotFoundError{reason: "no associated token"}
	}
	return userID, nil
}
