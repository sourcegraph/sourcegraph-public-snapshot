package billing

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// dbBilling provides billing-related database operations.
type dbBilling struct {
	db dbutil.DB
}

// getUserBillingCustomerID gets the billing customer ID (if any) for a user.
//
// If a transaction tx is provided, the query is executed using the transaction. If tx is nil, the
// global DB handle is used.
func (s dbBilling) getUserBillingCustomerID(ctx context.Context, userID int32) (billingCustomerID *string, err error) {
	query := sqlf.Sprintf("SELECT billing_customer_id FROM users WHERE id=%d AND deleted_at IS NULL", userID)
	err = s.db.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...).Scan(&billingCustomerID)
	if err == sql.ErrNoRows {
		return nil, database.NewUserNotFoundError(userID)
	}
	return billingCustomerID, err
}

// setUserBillingCustomerID sets or unsets the billing customer ID for a user.
//
// If a transaction tx is provided, the query is executed using the transaction. If tx is nil, the
// global DB handle is used.
func (s dbBilling) setUserBillingCustomerID(ctx context.Context, userID int32, billingCustomerID *string) error {
	query := sqlf.Sprintf("UPDATE users SET billing_customer_id=%s WHERE id=%d AND deleted_at IS NULL", billingCustomerID, userID)
	res, err := s.db.ExecContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return database.NewUserNotFoundError(userID)
	}
	return nil
}
