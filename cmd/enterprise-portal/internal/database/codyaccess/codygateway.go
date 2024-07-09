package codyaccess

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/internal/pgxerrors"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/internal/upsert"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type TableCodyGatewayAccess struct {
	// ⚠️ DO NOT USE: This field is only used for creating foreign key constraint.
	Subscription *subscriptions.TableSubscription `gorm:"foreignKey:SubscriptionID"`

	CodyGatewayAccess
}

func (s *TableCodyGatewayAccess) TableName() string {
	return "enterprise_portal_cody_gateway_access"
}

type CodyGatewayAccess struct {
	// SubscriptionID is the internal unprefixed UUID of the related subscription.
	SubscriptionID string `gorm:"type:uuid;not null;unique"`

	// Whether or not a subscription has Cody Gateway access enabled.
	Enabled bool `gorm:"not null"`

	// chat_completions_rate_limit
	ChatCompletionsRateLimit                *int64 `gorm:"type:bigint"`
	ChatCompletionsRateLimitIntervalSeconds *int

	// code_completions_rate_limit
	CodeCompletionsRateLimit                *int64 `gorm:"type:bigint"`
	CodeCompletionsRateLimitIntervalSeconds *int

	// embeddings_rate_limit
	EmbeddingsRateLimit                *int64 `gorm:"type:bigint"`
	EmbeddingsRateLimitIntervalSeconds *int
}

// codyGatewayAccessTableColumns must match scanCodyGatewayAccess() values.
// Requires:
//
//	RIGHT JOIN
//		enterprise_portal_subscriptions AS subscription
//		ON subscription_id = subscription.id
//
// As this uses subscription fields.
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
		"subscription.display_name",
	}
}

// scanCodyGatewayAccess matches s.columns() values.
func scanCodyGatewayAccess(row pgx.Row) (*CodyGatewayAccessWithSubscriptionDetails, error) {
	var a CodyGatewayAccessWithSubscriptionDetails
	// RIGHT JOIN may surface null in enterprise_portal_cody_gateway_access if
	// an active subscription exists, but explicit access is not configured. In
	// this case we still need to return a valid CodyGatewayAccessWithSubscriptionDetails,
	// just with empty fields.
	var maybeEnabled *bool
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
		&a.DisplayName,
	)
	if err != nil {
		return nil, err
	}
	if maybeEnabled != nil {
		a.Enabled = *maybeEnabled
	}
	return &a, nil
}

// Store is the storage layer for Cody Gateway access.
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
}

var ErrSubscriptionDoesNotExist = errors.New("subscription does not exist")

// Get returns the Cody Gateway access for the given subscription.
func (s *CodyGatewayStore) Get(ctx context.Context, subscriptionID string) (*CodyGatewayAccessWithSubscriptionDetails, error) {
	query := fmt.Sprintf(`SELECT
	%s
FROM
	enterprise_portal_cody_gateway_access
RIGHT JOIN
	enterprise_portal_subscriptions AS subscription
	ON subscription_id = subscription.id
WHERE
	subscription.id = @subscriptionID
	AND subscription.archived_at IS NULL`,
		strings.Join(codyGatewayAccessTableColumns(), ", "))

	sub, err := scanCodyGatewayAccess(s.db.QueryRow(ctx, query, pgx.NamedArgs{
		"subscriptionID": subscriptionID,
	}))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// RIGHT JOIN in query ensures that if we find no result, it's
			// because the subscription does not exist or is archived.
			return nil, errors.WithSafeDetails(
				errors.WithStack(ErrSubscriptionDoesNotExist),
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
	enterprise_portal_cody_gateway_access
RIGHT JOIN
	enterprise_portal_subscriptions AS subscription
	ON subscription_id = subscription.id
WHERE
	subscription.archived_at IS NULL`,
		strings.Join(codyGatewayAccessTableColumns(), ", "))

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
	Enabled bool

	// chat_completions_rate_limit
	ChatCompletionsRateLimit                *int64
	ChatCompletionsRateLimitIntervalSeconds *int

	// code_completions_rate_limit
	CodeCompletionsRateLimit                *int64
	CodeCompletionsRateLimitIntervalSeconds *int

	// embeddings_rate_limit
	EmbeddingsRateLimit                *int64
	EmbeddingsRateLimitIntervalSeconds *int

	// ForceUpdate indicates whether to force update all fields of the subscription
	// record.
	ForceUpdate bool
}

// toQuery returns the query based on the options. It returns an empty query if
// nothing to update.
func (opts UpsertCodyGatewayAccessOptions) Exec(ctx context.Context, db *pgxpool.Pool, subscriptionID string) error {
	b := upsert.New("enterprise_portal_cody_gateway_access", "subscription_id", opts.ForceUpdate)
	upsert.Field(b, "subscription_id", subscriptionID)
	upsert.Field(b, "enabled", opts.Enabled)
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
				errors.WithStack(ErrSubscriptionDoesNotExist),
				err.Error())
		}
		return nil, errors.Wrap(err, "exec")
	}
	return s.Get(ctx, subscriptionID)
}
