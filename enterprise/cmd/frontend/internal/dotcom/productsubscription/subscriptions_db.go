package productsubscription

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/dbconn"
)

// dbSubscription describes an product subscription row in the product_subscriptions DB
// table.
type dbSubscription struct {
	ID                    string // UUID
	UserID                int32
	BillingSubscriptionID *string // this subscription's ID in the billing system
	CreatedAt             time.Time
	ArchivedAt            *time.Time
}

// errSubscriptionNotFound occurs when a database operation expects a specific Sourcegraph
// license to exist but it does not exist.
var errSubscriptionNotFound = errors.New("product subscription not found")

// dbSubscriptions exposes product subscriptions in the product_subscriptions DB table.
type dbSubscriptions struct{}

// Create creates a new product subscription entry given a license key.
func (dbSubscriptions) Create(ctx context.Context, userID int32) (id string, err error) {
	if mocks.subscriptions.Create != nil {
		return mocks.subscriptions.Create(userID)
	}

	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	if err := dbconn.Global.QueryRowContext(ctx, `
INSERT INTO product_subscriptions(id, user_id) VALUES($1, $2) RETURNING id
`,
		uuid, userID,
	).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

// GetByID retrieves the product subscription (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this product subscription.
func (s dbSubscriptions) GetByID(ctx context.Context, id string) (*dbSubscription, error) {
	if mocks.subscriptions.GetByID != nil {
		return mocks.subscriptions.GetByID(id)
	}
	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%s", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errSubscriptionNotFound
	}
	return results[0], nil
}

// dbSubscriptionsListOptions contains options for listing product subscriptions.
type dbSubscriptionsListOptions struct {
	UserID          int32 // only list product subscriptions for this user
	IncludeArchived bool
	*db.LimitOffset
}

func (o dbSubscriptionsListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.UserID != 0 {
		conds = append(conds, sqlf.Sprintf("user_id=%d", o.UserID))
	}
	if !o.IncludeArchived {
		conds = append(conds, sqlf.Sprintf("archived_at IS NULL"))
	}
	return conds
}

// List lists all product subscriptions that satisfy the options.
func (s dbSubscriptions) List(ctx context.Context, opt dbSubscriptionsListOptions) ([]*dbSubscription, error) {
	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (dbSubscriptions) list(ctx context.Context, conds []*sqlf.Query, limitOffset *db.LimitOffset) ([]*dbSubscription, error) {
	q := sqlf.Sprintf(`
SELECT id, user_id, billing_subscription_id, created_at, archived_at FROM product_subscriptions
WHERE (%s)
ORDER BY archived_at DESC NULLS FIRST, created_at DESC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)

	rows, err := dbconn.Global.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*dbSubscription
	for rows.Next() {
		var v dbSubscription
		if err := rows.Scan(&v.ID, &v.UserID, &v.BillingSubscriptionID, &v.CreatedAt, &v.ArchivedAt); err != nil {
			return nil, err
		}
		results = append(results, &v)
	}
	return results, nil
}

// Count counts all product subscriptions that satisfy the options (ignoring limit and offset).
func (dbSubscriptions) Count(ctx context.Context, opt dbSubscriptionsListOptions) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM product_subscriptions WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// dbSubscriptionsUpdate represents an update to a product subscription in the database. Each field
// represents an update to the corresponding database field if the Go value is non-nil. If the Go
// value is nil, the field remains unchanged in the database.
type dbSubscriptionUpdate struct {
	billingSubscriptionID *sql.NullString
}

// Update updates a product subscription.
func (dbSubscriptions) Update(ctx context.Context, id string, update dbSubscriptionUpdate) error {
	fieldUpdates := []*sqlf.Query{
		sqlf.Sprintf("updated_at=now()"), // always update updated_at timestamp
	}
	if v := update.billingSubscriptionID; v != nil {
		fieldUpdates = append(fieldUpdates, sqlf.Sprintf("billing_subscription_id=%s", *v))
	}

	query := sqlf.Sprintf("UPDATE product_subscriptions SET %s WHERE id=%s", sqlf.Join(fieldUpdates, ", "), id)
	res, err := dbconn.Global.ExecContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errSubscriptionNotFound
	}
	return nil
}

// Archive marks a product subscription as archived given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to archive the token.
func (dbSubscriptions) Archive(ctx context.Context, id string) error {
	if mocks.subscriptions.Archive != nil {
		return mocks.subscriptions.Archive(id)
	}
	q := sqlf.Sprintf("UPDATE product_subscriptions SET archived_at=now(), updated_at=now() WHERE id=%s AND archived_at IS NULL", id)
	res, err := dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errSubscriptionNotFound
	}
	return nil
}

type mockSubscriptions struct {
	Create  func(userID int32) (id string, err error)
	GetByID func(id string) (*dbSubscription, error)
	Archive func(id string) error
}
