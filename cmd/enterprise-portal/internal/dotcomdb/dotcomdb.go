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
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	codyaccessv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/codyaccess/v1"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Reader struct {
	db   *pgxpool.Pool
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
func NewReader(db *pgxpool.Pool, opts ReaderOptions) *Reader {
	return &Reader{db: db, opts: opts}
}

func (r *Reader) Ping(ctx context.Context) error {
	// Execute ping steps within a single connection.
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "db.Acquire")
	}
	defer conn.Release()

	if err := conn.Ping(ctx); err != nil {
		return errors.Wrap(err, "sqlDB.PingContext")
	}
	if _, err := conn.Exec(ctx, "SELECT current_user;"); err != nil {
		return errors.Wrap(err, "sqlDB.Exec SELECT current_user")
	}
	return nil
}

func (r *Reader) Close() { r.db.Close() }

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

func (c CodyGatewayAccessAttributes) GetSubscriptionDisplayName() string {
	for _, tag := range c.ActiveLicenseTags {
		if strings.HasPrefix(tag, "customer:") {
			return strings.TrimPrefix(tag, "customer:")
		}
	}
	return ""
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
		Code:       licensing.NewCodyGatewayCodeRateLimit(p, c.ActiveLicenseUserCount),

		EmbeddingsSource: codyaccessv1.CodyGatewayRateLimitSource_CODY_GATEWAY_RATE_LIMIT_SOURCE_PLAN,
		Embeddings:       licensing.NewCodyGatewayEmbeddingsRateLimit(p, c.ActiveLicenseUserCount),
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
	accessTokens := make([]string, 0, len(c.LicenseKeyHashes))
	for _, t := range c.LicenseKeyHashes {
		if len(t) == 0 { // query can return empty hashes, ignore these
			continue
		}
		// See license.GenerateLicenseKeyBasedAccessToken
		accessTokens = append(accessTokens, license.LicenseKeyBasedAccessTokenPrefix+hex.EncodeToString(t))
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
		c := fmt.Sprintf("ARRAY['%s'] && licenses.license_tags",
			licensing.DevTag)
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

// ListEnterpriseSubscriptionLicenses returns a list of enterprise subscription
// license attributes with the given filters. It silently ignores any
// non-matching filters. The caller should check the length of the returned
// slice to ensure all requested licenses were found.
func (r *Reader) ListEnterpriseSubscriptionLicenses(
	ctx context.Context,
	filters []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter,
	pageSize int,
) ([]*LicenseAttributes, error) {
	conds := queryConditions{
		limit: pageSize,
	}
	var args []any
	var hasRevokedFilter bool
	for _, filter := range filters {
		switch filter.GetFilter().(type) {
		case *subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_SubscriptionId:
			conds.addWhere(fmt.Sprintf("licenses.product_subscription_id = $%d", len(args)+1))
			args = append(args,
				strings.TrimPrefix(filter.GetSubscriptionId(), subscriptionsv1.EnterpriseSubscriptionIDPrefix))
		case *subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter_IsRevoked:
			hasRevokedFilter = true
			// We treat subscription archived and revoked license the same.
			if filter.GetIsRevoked() {
				conds.addWhere("(subscriptions.archived_at IS NOT NULL OR licenses.revoked_at IS NOT NULL)")
			} else {
				conds.addWhere("subscriptions.archived_at IS NULL")
				conds.addWhere("licenses.revoked_at IS NULL")
			}
		}
	}
	if !hasRevokedFilter {
		conds.addWhere("subscriptions.archived_at IS NULL")
		conds.addWhere("licenses.revoked_at IS NULL")
	}

	query := newLicensesQuery(conds, r.opts)
	rows, err := r.db.Query(ctx, query, args...)
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

type SubscriptionAttributes struct {
	ID         string // UUID-format ID
	CreatedAt  time.Time
	ArchivedAt *time.Time

	UserDisplayName string
}

func (s SubscriptionAttributes) GenerateDisplayName() string {
	var parts []string
	if s.UserDisplayName != "" {
		parts = append(parts, s.UserDisplayName)
	}
	parts = append(parts,
		// Stick a seconds-granularity component to the name to guarantee
		// uniqueness during migration.
		s.CreatedAt.Format(time.DateTime))
	return strings.Join(parts, " - ")
}

type ListEnterpriseSubscriptionsOptions struct {
	SubscriptionIDs []string
	IsArchived      bool
}

// ListEnterpriseSubscriptions returns a list of enterprise subscription
// attributes with the given IDs. It silently ignores any non-existent
// subscription IDs. The caller should check the length of the returned slice to
// ensure all requested subscriptions were found.
//
// If no IDs are given, it returns all subscriptions.
func (r *Reader) ListEnterpriseSubscriptions(ctx context.Context, opts ListEnterpriseSubscriptionsOptions) ([]*SubscriptionAttributes, error) {
	query := `
SELECT
	product_subscriptions.id,
	product_subscriptions.created_at,
	product_subscriptions.archived_at,
	COALESCE( NULLIF(users.display_name, ''), users.username ) AS user_display_name
FROM
	product_subscriptions
JOIN users ON users.id = product_subscriptions.user_id
WHERE true`
	namedArgs := pgx.NamedArgs{}
	if len(opts.SubscriptionIDs) > 0 {
		query += "\nAND product_subscriptions.id = ANY(@ids)"
		namedArgs["ids"] = opts.SubscriptionIDs
	}
	if opts.IsArchived {
		query += "\nAND product_subscriptions.archived_at IS NOT NULL"
	} else {
		query += "\nAND product_subscriptions.archived_at IS NULL"
	}
	var licenseCond string
	if r.opts.DevOnly {
		licenseCond = fmt.Sprintf("'%s' = ANY(product_licenses.license_tags)", licensing.DevTag)
	} else {
		licenseCond = "true"
	}
	query += fmt.Sprintf(`
AND EXISTS (
	SELECT 1
	FROM product_licenses
	WHERE product_licenses.product_subscription_id = product_subscriptions.id
	AND %s
	ORDER BY product_licenses.created_at DESC
	LIMIT 1
)
`, licenseCond)

	rows, err := r.db.Query(ctx, query+"ORDER BY product_subscriptions.created_at DESC", namedArgs)
	if err != nil {
		return nil, errors.Wrap(err, "query subscription attributes")
	}
	defer rows.Close()
	var attrs []*SubscriptionAttributes
	for rows.Next() {
		var attr SubscriptionAttributes
		err = rows.Scan(&attr.ID, &attr.CreatedAt, &attr.ArchivedAt, &attr.UserDisplayName)
		if err != nil {
			return nil, errors.Wrap(err, "scan subscription attributes")
		}
		attrs = append(attrs, &attr)
	}
	return attrs, rows.Err()
}
