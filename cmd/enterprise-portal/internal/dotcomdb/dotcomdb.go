// Package dotcomdb provides a read-only shim over the Sourcegraph.com database
// and aims to provide values as they behave in Sourcegraph.com API today for
// Enterprise Portal to serve through its new API.
//
// 👷 This package is intended to be a short-lived mechanism, and should be
// removed as part of https://linear.app/sourcegraph/project/12f1d5047bd2/overview.
package dotcomdb

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/productsubscription"
	codyaccessv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/codyaccess/v1"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Reader struct {
	conn *pgx.Conn
	opts ReaderOptions
}

type ReaderOptions struct {
	// DevOnly indicates that this Reader should only return subscriptions,
	// licenses, etc. that are only used for development.
	DevOnly bool
}

// NewReader wraps a direct connection to the Sourcegraph.com database. It
// ONLY executes read queries, so the connection can (and should) be
// authenticated by a read-only user.
//
// 👷 This is intended to be a short-lived mechanism, and should be removed
// as part of https://linear.app/sourcegraph/project/12f1d5047bd2/overview.
func NewReader(conn *pgx.Conn, opts ReaderOptions) *Reader {
	return &Reader{conn: conn, opts: opts}
}

func (r *Reader) Ping(ctx context.Context) error {
	if err := r.conn.Ping(ctx); err != nil {
		return errors.Wrap(err, "sqlDB.PingContext")
	}
	if _, err := r.conn.Exec(ctx, "SELECT current_user;"); err != nil {
		return errors.Wrap(err, "sqlDB.Exec SELECT current_user")
	}
	return nil
}

func (r *Reader) Close(ctx context.Context) error { return r.conn.Close(ctx) }

type CodyGatewayAccessAttributes struct {
	SubscriptionID string

	CodyGatewayEnabled bool

	ChatCompletionsRateLimit   *int64
	ChatCompletionsRateSeconds *int32

	CodeCompletionsRateLimit   *int64
	CodeCompletionsRateSeconds *int32

	EmbeddingsRateLimit   *int64
	EmbeddingsRateSeconds *int32

	ActiveLicenseTags      []string
	ActiveLicenseUserCount *int

	// Used for GenerateAccessTokens
	LicenseKeyHashes [][]byte
}

type CodyGatewayRateLimits struct {
	ChatSource codyaccessv1.CodyGatewayRateLimitSource
	Chat       licensing.CodyGatewayRateLimit

	CodeSource codyaccessv1.CodyGatewayRateLimitSource
	Code       licensing.CodyGatewayRateLimit

	EmbeddingsSource codyaccessv1.CodyGatewayRateLimitSource
	Embeddings       licensing.CodyGatewayRateLimit
}

func maybeApplyOverride[T ~int32 | ~int64](limit *T, override *T) codyaccessv1.CodyGatewayRateLimitSource {
	if override != nil {
		*limit = *override
		return codyaccessv1.CodyGatewayRateLimitSource_CODY_GATEWAY_RATE_LIMIT_SOURCE_OVERRIDE
	}
	// No override
	return codyaccessv1.CodyGatewayRateLimitSource_CODY_GATEWAY_RATE_LIMIT_SOURCE_PLAN
}

// EvaluateRateLimits returns the current CodyGatewayRateLimits based on the
// plan and applying known overrides on top. This closely models the existing
// codyGatewayAccessResolver in 'cmd/frontend/internal/dotcom/productsubscription'.
func (c CodyGatewayAccessAttributes) EvaluateRateLimits() CodyGatewayRateLimits {
	// Set defaults for everything based on plan
	p := licensing.PlanFromTags(c.ActiveLicenseTags)
	limits := CodyGatewayRateLimits{
		ChatSource: codyaccessv1.CodyGatewayRateLimitSource_CODY_GATEWAY_RATE_LIMIT_SOURCE_PLAN,
		Chat:       licensing.NewCodyGatewayChatRateLimit(p, c.ActiveLicenseUserCount),

		CodeSource: codyaccessv1.CodyGatewayRateLimitSource_CODY_GATEWAY_RATE_LIMIT_SOURCE_PLAN,
		Code:       licensing.NewCodyGatewayCodeRateLimit(p, c.ActiveLicenseUserCount, c.ActiveLicenseTags),

		EmbeddingsSource: codyaccessv1.CodyGatewayRateLimitSource_CODY_GATEWAY_RATE_LIMIT_SOURCE_PLAN,
		Embeddings:       licensing.NewCodyGatewayEmbeddingsRateLimit(p, c.ActiveLicenseUserCount, c.ActiveLicenseTags),
	}

	// Chat
	limits.ChatSource = maybeApplyOverride(&limits.Chat.Limit, c.ChatCompletionsRateLimit)
	limits.ChatSource = maybeApplyOverride(&limits.Chat.IntervalSeconds, c.ChatCompletionsRateSeconds)

	// Code
	limits.CodeSource = maybeApplyOverride(&limits.Code.Limit, c.CodeCompletionsRateLimit)
	limits.CodeSource = maybeApplyOverride(&limits.Code.IntervalSeconds, c.CodeCompletionsRateSeconds)

	// Embeddings
	limits.EmbeddingsSource = maybeApplyOverride(&limits.Embeddings.Limit, c.EmbeddingsRateLimit)
	limits.EmbeddingsSource = maybeApplyOverride(&limits.Embeddings.IntervalSeconds, c.EmbeddingsRateSeconds)

	return limits
}

func (c CodyGatewayAccessAttributes) GenerateAccessTokens() []string {
	accessTokens := make([]string, len(c.LicenseKeyHashes))
	for i, t := range c.LicenseKeyHashes {
		// See license.GenerateLicenseKeyBasedAccessToken
		accessTokens[i] = license.LicenseKeyBasedAccessTokenPrefix + hex.EncodeToString(t)
	}
	return accessTokens
}

var ErrCodyGatewayAccessNotFound = errors.New("cody gateway access not found")

// Clauses can reference 'subscription', 'tokens', and 'active_license'.
type queryConditions struct {
	whereClause  string
	havingClause string
	limit        int
}

func (q *queryConditions) addWhere(cond string) {
	if q.whereClause != "" {
		q.whereClause += " AND " + cond
	} else {
		q.whereClause = cond
	}
}

func newCodyGatewayAccessQuery(conds queryConditions, opts ReaderOptions) string {
	const rawClause = `
SELECT
	subscription.id,
	subscription.cody_gateway_enabled,
	-- ChatCompletionsRateLimit override
	subscription.cody_gateway_chat_rate_limit,
	subscription.cody_gateway_chat_rate_interval_seconds,
	-- CodeCompletionsRateLimit override
	subscription.cody_gateway_code_rate_limit,
	subscription.cody_gateway_code_rate_interval_seconds,
	-- EmbeddingsRateLimit override
	subscription.cody_gateway_embeddings_api_rate_limit,
	subscription.cody_gateway_embeddings_api_rate_interval_seconds,
	-- "Active license": we aggregate for tokens below, so we need to apply MAX
	-- here to make this look like an aggregated value. This is okay becuase
	-- active_license uses 'SELECT DISTINCT ON' which returns exactly 1 row.
	MAX(active_license.license_tags),
	MAX(active_license.license_user_count),
	-- All past license keys that can be used as "access tokens"
	array_agg(tokens.license_key_hash) as license_key_hashes
FROM product_subscriptions subscription
	LEFT JOIN (
		SELECT DISTINCT ON (licenses.product_subscription_id)
			licenses.product_subscription_id,
			licenses.license_tags,
			licenses.license_user_count
		FROM product_licenses AS licenses
		-- Get most recently created license key as the "active license"
		ORDER BY licenses.product_subscription_id, licenses.created_at DESC
	) active_license ON active_license.product_subscription_id = subscription.id
	LEFT JOIN (
		SELECT
			licenses.product_subscription_id,
			digest(licenses.license_key, 'sha256') AS license_key_hash
		FROM product_licenses as licenses
		WHERE licenses.access_token_enabled IS TRUE
	) tokens ON tokens.product_subscription_id = subscription.id`

	clauses := []string{rawClause}
	// Add WHERE clause, amending it to include a condition that the subscription
	// must not be archived.
	if conds.whereClause != "" {
		clauses = append(clauses, "WHERE "+conds.whereClause+" AND subscription.archived_at IS NULL")
	} else {
		clauses = append(clauses, "WHERE subscription.archived_at IS NULL")
	}
	clauses = append(clauses, "GROUP BY subscription.id") // required, after WHERE clause
	if conds.havingClause != "" {
		clauses = append(clauses, "HAVING "+conds.havingClause)
	}
	if opts.DevOnly {
		// '&&' operator: overlap (have elements in common)
		c := fmt.Sprintf("ARRAY['%s','%s'] && MAX(active_license.license_tags)",
			licensing.DevTag, licensing.InternalTag)
		if conds.havingClause != "" {
			clauses = append(clauses, "AND "+c)
		} else {
			clauses = append(clauses, "HAVING "+c)
		}
	}
	return strings.Join(clauses, "\n")
}

type GetCodyGatewayAccessAttributesOpts struct {
	BySubscription *string
	ByAccessToken  *string
}

func (r *Reader) GetCodyGatewayAccessAttributesBySubscription(ctx context.Context, subscriptionID string) (*CodyGatewayAccessAttributes, error) {
	query := newCodyGatewayAccessQuery(queryConditions{
		whereClause: "subscription.id = $1",
	}, r.opts)
	row := r.conn.QueryRow(ctx, query,
		strings.TrimPrefix(subscriptionID, subscriptionsv1.EnterpriseSubscriptionIDPrefix))
	return scanCodyGatewayAccessAttributes(row)
}

func (r *Reader) GetCodyGatewayAccessAttributesByAccessToken(ctx context.Context, token string) (*CodyGatewayAccessAttributes, error) {
	// Below is copied from 'func (t dbTokens) LookupProductSubscriptionIDByAccessToken'
	// in 'cmd/frontend/internal/dotcom/productsubscription'.
	if !strings.HasPrefix(token, productsubscription.AccessTokenPrefix) &&
		!strings.HasPrefix(token, license.LicenseKeyBasedAccessTokenPrefix) {
		return nil, errors.WithSafeDetails(ErrCodyGatewayAccessNotFound, "invalid token with unknown prefix")
	}
	tokenSansPrefix := token[len(license.LicenseKeyBasedAccessTokenPrefix):]
	decoded, err := hex.DecodeString(tokenSansPrefix)
	if err != nil {
		return nil, errors.WithSafeDetails(ErrCodyGatewayAccessNotFound, "invalid token with unknown encoding")
	}
	// End copied code.

	query := newCodyGatewayAccessQuery(queryConditions{
		havingClause: "$1 = ANY(array_agg(tokens.license_key_hash))",
	}, r.opts)
	row := r.conn.QueryRow(ctx, query, decoded)
	return scanCodyGatewayAccessAttributes(row)
}

func scanCodyGatewayAccessAttributes(row pgx.Row) (*CodyGatewayAccessAttributes, error) {
	var attrs CodyGatewayAccessAttributes
	err := row.Scan(
		&attrs.SubscriptionID,
		&attrs.CodyGatewayEnabled,
		&attrs.ChatCompletionsRateLimit,
		&attrs.ChatCompletionsRateSeconds,
		&attrs.CodeCompletionsRateLimit,
		&attrs.CodeCompletionsRateSeconds,
		&attrs.EmbeddingsRateLimit,
		&attrs.EmbeddingsRateSeconds,
		&attrs.ActiveLicenseTags,
		&attrs.ActiveLicenseUserCount,
		&attrs.LicenseKeyHashes,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.WithStack(ErrCodyGatewayAccessNotFound)
		}
		return nil, errors.Wrap(err, "failed to get cody gateway access attributes")
	}
	return &attrs, nil
}

func (r *Reader) GetAllCodyGatewayAccessAttributes(ctx context.Context) ([]*CodyGatewayAccessAttributes, error) {
	query := newCodyGatewayAccessQuery(queryConditions{}, r.opts)
	rows, err := r.conn.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cody gateway access attributes")
	}
	defer rows.Close()
	var attrs []*CodyGatewayAccessAttributes
	for rows.Next() {
		attr, err := scanCodyGatewayAccessAttributes(rows)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan cody gateway access attributes")
		}
		attrs = append(attrs, attr)
	}
	return attrs, rows.Err()
}

var ErrEnterpriseSubscriptionLicenseNotFound = errors.New("enterprise subscription license not found")

func newLicensesQuery(conds queryConditions, opts ReaderOptions) string {
	const rawClause = `
SELECT
	-- EnterpriseSubscriptionLicense
	licenses.id,
	licenses.product_subscription_id,
	-- EnterpriseSubscriptionLicenseCondition
	licenses.created_at,
	licenses.revoked_at,
	licenses.revoke_reason,
	-- EnterpriseSubscriptionLicenseKey
	licenses.license_version,
	licenses.license_tags,
	licenses.license_user_count,
	licenses.license_expires_at,
	licenses.salesforce_sub_id,
	licenses.salesforce_opp_id,
	licenses.license_key,
	licenses.site_id
FROM product_licenses licenses
LEFT JOIN product_subscriptions subscriptions
	ON subscriptions.id = licenses.product_subscription_id
`
	clauses := []string{rawClause}
	if conds.whereClause != "" {
		clauses = append(clauses, "WHERE "+conds.whereClause)
	}
	if opts.DevOnly {
		// '&&' operator: overlap (have elements in common)
		c := fmt.Sprintf("ARRAY['%s','%s'] && licenses.license_tags",
			licensing.DevTag, licensing.InternalTag)
		if conds.whereClause != "" {
			clauses = append(clauses, "AND "+c)
		} else {
			clauses = append(clauses, "WHERE "+c)
		}
	}
	if conds.havingClause != "" {
		clauses = append(clauses, "HAVING "+conds.havingClause)
	}
	clauses = append(clauses, "ORDER BY licenses.created_at DESC")
	if conds.limit > 0 {
		clauses = append(clauses, fmt.Sprintf("LIMIT %d", conds.limit))
	}
	return strings.Join(clauses, "\n")
}

type LicenseAttributes struct {
	// EnterpriseSubscriptionLicense
	ID             string
	SubscriptionID string
	// EnterpriseSubscriptionLicenseCondition
	CreatedAt    time.Time
	RevokedAt    *time.Time
	RevokeReason *string
	// EnterpriseSubscriptionLicenseKey
	InfoVersion              *uint32
	Tags                     []string
	UserCount                *uint64
	ExpiresAt                *time.Time
	SalesforceSubscriptionID *string
	SalesforceOpportunityID  *string
	LicenseKey               string
	InstanceID               *string
}

func scanLicenseAttributes(row pgx.Row) (*LicenseAttributes, error) {
	var attrs LicenseAttributes
	err := row.Scan(
		// EnterpriseSubscriptionLicense
		&attrs.ID,
		&attrs.SubscriptionID,
		// EnterpriseSubscriptionLicenseCondition
		&attrs.CreatedAt,
		&attrs.RevokedAt,
		&attrs.RevokeReason,
		// EnterpriseSubscriptionLicenseKey
		&attrs.InfoVersion,
		&attrs.Tags,
		&attrs.UserCount,
		&attrs.ExpiresAt,
		&attrs.SalesforceSubscriptionID,
		&attrs.SalesforceOpportunityID,
		&attrs.LicenseKey,
		&attrs.InstanceID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.WithStack(ErrEnterpriseSubscriptionLicenseNotFound)
		}
		return nil, errors.Wrap(err, "failed to get enterprise subscription license attributes")
	}
	return &attrs, nil
}

func (r *Reader) ListEnterpriseSubscriptionLicenses(
	ctx context.Context,
	filters []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter,
	pageSize int,
) ([]*LicenseAttributes, error) {
	conds := queryConditions{
		limit: pageSize,
	}
	var args []any
	for _, filter := range filters {
		switch filter.GetFilter().(type) {
		case *subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_SubscriptionId:
			conds.addWhere(fmt.Sprintf("licenses.product_subscription_id = $%d", len(args)+1))
			args = append(args,
				strings.TrimPrefix(filter.GetSubscriptionId(), subscriptionsv1.EnterpriseSubscriptionIDPrefix))
		case *subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_IsArchived:
			if filter.GetIsArchived() {
				conds.addWhere("subscriptions.archived_at IS NOT NULL")
			} else {
				conds.addWhere("subscriptions.archived_at IS NULL")
			}
		}
	}

	query := newLicensesQuery(conds, r.opts)
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cody gateway access attributes")
	}
	defer rows.Close()
	var attrs []*LicenseAttributes
	for rows.Next() {
		attr, err := scanLicenseAttributes(rows)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan cody gateway access attributes")
		}
		attrs = append(attrs, attr)
	}
	return attrs, rows.Err()
}
