package codyaccess

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/internal/pgxerrors"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/internal/upsert"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ⚠️ DO NOT USE: This type is only used for creating foreign key constraints
// and initializing tables with gorm.
type TableCodyGatewayAccess struct {
	Subscription *subscriptions.TableSubscription `gorm:"foreignKey:SubscriptionID"`

	CodyGatewayAccess
}

func (*TableCodyGatewayAccess) TableName() string {
	return "enterprise_portal_cody_gateway_access"
}

func (t *TableCodyGatewayAccess) RunCustomMigrations(migrator gorm.Migrator) error {
	// gorm seems to refuse to drop the 'not null' constriant on a column
	// unless we forcibly run AlterColumn.
	columns := []string{
		"chat_completions_rate_limit",
		"chat_completions_rate_limit_interval_seconds",
		"code_completions_rate_limit",
		"code_completions_rate_limit_interval_seconds",
		"embeddings_rate_limit",
		"embeddings_rate_limit_interval_seconds",
	}
	for _, column := range columns {
		if err := migrator.AlterColumn(t, column); err != nil {
			return err
		}
	}
	return nil
}

type CodyGatewayAccess struct {
	// SubscriptionID is the internal unprefixed UUID of the related subscription.
	SubscriptionID string `gorm:"type:uuid;not null;unique"`

	// Whether or not a subscription has Cody Gateway access enabled.
	Enabled bool `gorm:"not null;default:false"`

	// chat_completions_rate_limit
	ChatCompletionsRateLimit                sql.NullInt64
	ChatCompletionsRateLimitIntervalSeconds sql.NullInt32

	// code_completions_rate_limit
	CodeCompletionsRateLimit                sql.NullInt64
	CodeCompletionsRateLimitIntervalSeconds sql.NullInt32

	// embeddings_rate_limit
	EmbeddingsRateLimit                sql.NullInt64
	EmbeddingsRateLimitIntervalSeconds sql.NullInt32
}

// codyGatewayAccessTableColumns must match scanCodyGatewayAccess() values.
// Requires 'codyGatewayAccessJoinClauses'.
func codyGatewayAccessTableColumns() []string {
	return []string{
		"subscription.id",
		"enabled",
		"chat_completions_rate_limit",
		"chat_completions_rate_limit_interval_seconds",
		"code_completions_rate_limit",
		"code_completions_rate_limit_interval_seconds",
		"embeddings_rate_limit",
		"embeddings_rate_limit_interval_seconds",
		// Subscriptions
		"subscription.display_name",
		// Licenses - depends on license key info
		"active_license.license_data->'Info' as active_license_info",
		"tokens.license_key_hashes as license_key_hashes",
	}
}

// scanCodyGatewayAccess matches s.columns() values.
func scanCodyGatewayAccess(row pgx.Row) (*CodyGatewayAccessWithSubscriptionDetails, error) {
	var a CodyGatewayAccessWithSubscriptionDetails
	// RIGHT JOIN may surface null in enterprise_portal_cody_gateway_access if
	// an active subscription exists, but explicit access is not configured. In
	// this case we still need to return a valid CodyGatewayAccessWithSubscriptionDetails,
	// just with empty fields.
	var maybeEnabled sql.NullBool
	var maybeDisplayName sql.NullString
	err := row.Scan(
		&a.SubscriptionID,
		&maybeEnabled,
		&a.ChatCompletionsRateLimit,
		&a.ChatCompletionsRateLimitIntervalSeconds,
		&a.CodeCompletionsRateLimit,
		&a.CodeCompletionsRateLimitIntervalSeconds,
		&a.EmbeddingsRateLimit,
		&a.EmbeddingsRateLimitIntervalSeconds,
		// Subscriptions fields
		&maybeDisplayName,
		// License fields
		&a.ActiveLicenseInfo,
		&a.LicenseKeyHashes,
	)
	if err != nil {
		return nil, err
	}
	a.Enabled = maybeEnabled.Bool
	a.DisplayName = maybeDisplayName.String

	return &a, nil
}

const codyGatewayAccessJoinClauses = `
-- We want Cody Gateway access records for every subscription, even if an
-- an explicit one doesn't exist yet.
RIGHT JOIN
    enterprise_portal_subscriptions AS subscription
    ON access.subscription_id = subscription.id

-- Join against the "active license" of a subscription, which is currently used
-- as the source for default subscription access properties.
-- We may want to move user counts, product tags, etc. to the subscription table
-- in the future instead.
LEFT JOIN
    enterprise_portal_subscription_licenses AS active_license
    ON active_license.id = (
        SELECT id
        FROM enterprise_portal_subscription_licenses
        WHERE
            enterprise_portal_subscription_licenses.license_type = 'ENTERPRISE_SUBSCRIPTION_LICENSE_TYPE_KEY'
            AND access.subscription_id = enterprise_portal_subscription_licenses.subscription_id
			AND enterprise_portal_subscription_licenses.expire_at > NOW()  -- expires in future
			AND enterprise_portal_subscription_licenses.revoked_at IS NULL -- not revoked
        -- Get most recently created license key as the "active license"
        ORDER BY enterprise_portal_subscription_licenses.created_at DESC
        LIMIT 1
    )

-- Join against collected license key hashes of each subscription, which we use
-- as 'access tokens' to Cody Gateway
LEFT JOIN (
	SELECT
		licenses.subscription_id,
		ARRAY_AGG(digest(licenses.license_data->>'SignedKey','sha256')) AS license_key_hashes
	FROM
		enterprise_portal_subscription_licenses AS licenses
	WHERE
		licenses.license_type = 'ENTERPRISE_SUBSCRIPTION_LICENSE_TYPE_KEY'
		AND licenses.expire_at > NOW()  -- expires in future
		AND licenses.revoked_at IS NULL -- is not revoked
    GROUP BY
        licenses.subscription_id
) tokens ON tokens.subscription_id = subscription.id
`

// Store is the storage layer for Cody Gateway access. It aims to mirror the
// existing behaviour as close as possible, and as such has extensive
// dependencies on licensing.
type CodyGatewayStore struct {
	db *pgxpool.Pool
}

func NewCodyGatewayStore(db *pgxpool.Pool) *CodyGatewayStore {
	return &CodyGatewayStore{db: db}
}

// CodyGatewayAccessWithSubscriptionDetails extends CodyGatewayAccess with metadata from
// other tables used in the codyaccess API.
type CodyGatewayAccessWithSubscriptionDetails struct {
	CodyGatewayAccess

	// DisplayName is the display name of the related subscription.
	DisplayName string

	ActiveLicenseInfo *license.Info

	// Used by GenerateAccessTokens
	LicenseKeyHashes [][]byte
}

var ErrSubscriptionNotFound = errors.New("subscription does not exist or is not valid for Cody Gateway access")

type GetCodyGatewayAccessOptions struct {
	SubscriptionID string
	LicenseKeyHash []byte
}

func (opts GetCodyGatewayAccessOptions) buildConds() (string, pgx.NamedArgs, error) {
	if opts.SubscriptionID == "" && len(opts.LicenseKeyHash) == 0 {
		return "", nil, errors.New("must specify either SubscriptionID or LicenseKeyHash")
	}

	args := pgx.NamedArgs{}
	conds := []string{"TRUE"}
	if opts.SubscriptionID != "" {
		conds = append(conds, "subscription.id = @subscriptionID")
		args["subscriptionID"] = opts.SubscriptionID
	}
	if len(opts.LicenseKeyHash) > 0 {
		conds = append(conds, "@licenseKeyHash = ANY(tokens.license_key_hashes)")
		args["licenseKeyHash"] = opts.LicenseKeyHash
	}
	return strings.Join(conds, " AND "), args, nil
}

// Get returns the Cody Gateway access for the given subscription.
func (s *CodyGatewayStore) Get(ctx context.Context, opts GetCodyGatewayAccessOptions) (*CodyGatewayAccessWithSubscriptionDetails, error) {
	conds, args, err := opts.buildConds()
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`SELECT
	%s
FROM
	enterprise_portal_cody_gateway_access AS access
%s
WHERE
	%s
	AND subscription.archived_at IS NULL`,
		strings.Join(codyGatewayAccessTableColumns(), ", "),
		codyGatewayAccessJoinClauses,
		conds)

	sub, err := scanCodyGatewayAccess(s.db.QueryRow(ctx, query, args))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// RIGHT JOIN in query ensures that if we find no result, it's
			// because the subscription does not exist or is archived.
			return nil, errors.WithSafeDetails(
				errors.WithStack(ErrSubscriptionNotFound),
				err.Error())
		}
		return nil, err
	}
	return sub, nil
}

func (s *CodyGatewayStore) List(ctx context.Context) ([]*CodyGatewayAccessWithSubscriptionDetails, error) {
	query := fmt.Sprintf(`SELECT
	%s
FROM
	enterprise_portal_cody_gateway_access AS access
%s
WHERE
	subscription.archived_at IS NULL`,
		strings.Join(codyGatewayAccessTableColumns(), ", "),
		codyGatewayAccessJoinClauses)

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var accs []*CodyGatewayAccessWithSubscriptionDetails
	for rows.Next() {
		sub, err := scanCodyGatewayAccess(rows)
		if err != nil {
			return nil, err
		}
		accs = append(accs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accs, nil
}

type UpsertCodyGatewayAccessOptions struct {
	// Whether or not a subscription has Cody Gateway access enabled.
	Enabled *bool

	// chat_completions_rate_limit
	ChatCompletionsRateLimit                *sql.NullInt64
	ChatCompletionsRateLimitIntervalSeconds *sql.NullInt32

	// code_completions_rate_limit
	CodeCompletionsRateLimit                *sql.NullInt64
	CodeCompletionsRateLimitIntervalSeconds *sql.NullInt32

	// embeddings_rate_limit
	EmbeddingsRateLimit                *sql.NullInt64
	EmbeddingsRateLimitIntervalSeconds *sql.NullInt32

	// ForceUpdate indicates whether to force update all fields of the subscription
	// record.
	ForceUpdate bool
}

// toQuery returns the query based on the options. It returns an empty query if
// nothing to update.
func (opts UpsertCodyGatewayAccessOptions) Exec(ctx context.Context, db *pgxpool.Pool, subscriptionID string) error {
	b := upsert.New("enterprise_portal_cody_gateway_access", "subscription_id", opts.ForceUpdate)
	upsert.Field(b, "subscription_id", subscriptionID)
	upsert.Field(b, "enabled", opts.Enabled,
		upsert.WithColumnDefault(),
		upsert.WithValueOnForceUpdate(false))
	upsert.Field(b, "chat_completions_rate_limit", opts.ChatCompletionsRateLimit)
	upsert.Field(b, "chat_completions_rate_limit_interval_seconds", opts.ChatCompletionsRateLimitIntervalSeconds)
	upsert.Field(b, "code_completions_rate_limit", opts.CodeCompletionsRateLimit)
	upsert.Field(b, "code_completions_rate_limit_interval_seconds", opts.CodeCompletionsRateLimitIntervalSeconds)
	upsert.Field(b, "embeddings_rate_limit", opts.EmbeddingsRateLimit)
	upsert.Field(b, "embeddings_rate_limit_interval_seconds", opts.EmbeddingsRateLimitIntervalSeconds)
	return b.Exec(ctx, db)
}

// Upsert upserts a Cody Gatweway access record based on the given options.
// The caller should check that the subscription is not archived.
//
// If the subscription does not exist, then ErrSubscriptionDoesNotExist is
// returned.
func (s *CodyGatewayStore) Upsert(ctx context.Context, subscriptionID string, opts UpsertCodyGatewayAccessOptions) (*CodyGatewayAccessWithSubscriptionDetails, error) {
	if err := opts.Exec(ctx, s.db, subscriptionID); err != nil {
		if pgxerrors.IsContraintError(err, "fk_enterprise_portal_cody_gateway_access_subscription") {
			return nil, errors.WithSafeDetails(
				errors.WithStack(ErrSubscriptionNotFound),
				err.Error())
		}
		return nil, err
	}
	return s.Get(ctx, GetCodyGatewayAccessOptions{SubscriptionID: subscriptionID})
}
