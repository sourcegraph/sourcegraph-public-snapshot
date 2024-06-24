package subscriptions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Subscription is an Enterprise subscription record.
type Subscription struct {
	// ID is the internal (unprefixed) UUID-format identifier for the subscription.
	ID string `gorm:"type:uuid;primaryKey"`
	// InstanceDomain is the instance domain associated with the subscription, e.g.
	// "acme.sourcegraphcloud.com". This is set explicitly.
	//
	// It must be unique across all currently un-archived subscriptions.
	InstanceDomain string `gorm:"uniqueIndex:,where:archived_at IS NULL"`

	// WARNING: The below fields are not yet used in production.

	// DisplayName is the human-friendly name of this subscription, e.g. "Acme, Inc."
	//
	// It must be unique across all currently un-archived subscriptions.
	DisplayName string `gorm:"size:256;not null;uniqueIndex:,where:archived_at IS NULL;default:'Unnamed subscription'"`

	// Timestamps representing the latest timestamps of key conditions related
	// to this subscription.
	//
	// Condition transition details are tracked in 'enterprise_portal_subscription_conditions'.
	CreatedAt  time.Time  `gorm:"not null;default:current_timestamp"`
	UpdatedAt  time.Time  `gorm:"not null;default:current_timestamp"`
	ArchivedAt *time.Time // Null indicates the subscription is not archived.

	// SalesforceSubscriptionID associated with this Enterprise subscription.
	SalesforceSubscriptionID *string
	// SalesforceOpportunityID associated with this Enterprise subscription.
	SalesforceOpportunityID *string
}

func (s *Subscription) TableName() string {
	return "enterprise_portal_subscriptions"
}

// Store is the storage layer for product subscriptions.
type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		db: db,
	}
}

// ListEnterpriseSubscriptionsOptions is the set of options to filter subscriptions.
// Non-empty fields are treated as AND-concatenated.
type ListEnterpriseSubscriptionsOptions struct {
	// IDs is a list of subscription IDs to filter by.
	IDs []string
	// InstanceDomains is a list of instance domains to filter by.
	InstanceDomains []string
	// IsArchived indicates whether to only list archived subscriptions, or only
	// non-archived subscriptions.
	IsArchived bool
	// PageSize is the maximum number of subscriptions to return.
	PageSize int
}

func (opts ListEnterpriseSubscriptionsOptions) toQueryConditions() (where, limit string, _ pgx.NamedArgs) {
	whereConds := []string{"TRUE"}
	namedArgs := pgx.NamedArgs{}
	if len(opts.IDs) > 0 {
		whereConds = append(whereConds, "id = ANY(@ids)")
		namedArgs["ids"] = opts.IDs
	}
	if len(opts.InstanceDomains) > 0 {
		whereConds = append(whereConds, "instance_domain = ANY(@instanceDomains)")
		namedArgs["instanceDomains"] = opts.InstanceDomains
	}
	// Future: Uncomment the following block when the archived field is added to the table.
	// if opts.OnlyArchived {
	// whereConds = append(whereConds, "archived = TRUE")
	// }
	where = strings.Join(whereConds, " AND ")

	if opts.PageSize > 0 {
		limit = "LIMIT @pageSize"
		namedArgs["pageSize"] = opts.PageSize
	}
	return where, limit, namedArgs
}

// List returns a list of subscriptions based on the given options.
func (s *Store) List(ctx context.Context, opts ListEnterpriseSubscriptionsOptions) ([]*Subscription, error) {
	where, limit, namedArgs := opts.toQueryConditions()
	query := fmt.Sprintf(`
SELECT
	id,
	instance_domain
FROM enterprise_portal_subscriptions
WHERE %s
%s`,
		where, limit,
	)
	rows, err := s.db.Query(ctx, query, namedArgs)
	if err != nil {
		return nil, errors.Wrap(err, "query rows")
	}
	defer rows.Close()

	var subscriptions []*Subscription
	for rows.Next() {
		var subscription Subscription
		if err = rows.Scan(&subscription.ID, &subscription.InstanceDomain); err != nil {
			return nil, errors.Wrap(err, "scan row")
		}
		subscriptions = append(subscriptions, &subscription)
	}
	return subscriptions, rows.Err()
}

type UpsertSubscriptionOptions struct {
	InstanceDomain string
	// ForceUpdate indicates whether to force update all fields of the subscription
	// record.
	ForceUpdate bool
}

// toQuery returns the query based on the options. It returns an empty query if
// nothing to update.
func (opts UpsertSubscriptionOptions) toQuery(id string) (query string, _ pgx.NamedArgs) {
	const queryFmt = `
INSERT INTO enterprise_portal_subscriptions (id, instance_domain)
VALUES (@id, @instanceDomain)
ON CONFLICT (id)
DO UPDATE SET
	%s`
	namedArgs := pgx.NamedArgs{
		"id":             id,
		"instanceDomain": opts.InstanceDomain,
	}

	var sets []string
	if opts.ForceUpdate || opts.InstanceDomain != "" {
		sets = append(sets, "instance_domain = excluded.instance_domain")
	}
	if len(sets) == 0 {
		return "", nil
	}
	query = fmt.Sprintf(
		queryFmt,
		strings.Join(sets, ", "),
	)
	return query, namedArgs
}

// Upsert upserts a subscription record based on the given options.
func (s *Store) Upsert(ctx context.Context, subscriptionID string, opts UpsertSubscriptionOptions) (*Subscription, error) {
	query, namedArgs := opts.toQuery(subscriptionID)
	if query != "" {
		_, err := s.db.Exec(ctx, query, namedArgs)
		if err != nil {
			return nil, errors.Wrap(err, "exec")
		}
	}
	return s.Get(ctx, subscriptionID)
}

// Get returns a subscription record with the given subscription ID. It returns
// pgx.ErrNoRows if no such subscription exists.
func (s *Store) Get(ctx context.Context, subscriptionID string) (*Subscription, error) {
	var subscription Subscription
	query := `SELECT id, instance_domain FROM enterprise_portal_subscriptions WHERE id = @id`
	namedArgs := pgx.NamedArgs{"id": subscriptionID}
	err := s.db.QueryRow(ctx, query, namedArgs).Scan(&subscription.ID, &subscription.InstanceDomain)
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}
