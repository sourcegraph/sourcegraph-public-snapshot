package db

import (
	"context"
	"database/sql"
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// userExternalAccountNotFoundError is the error that is returned when a user external account is not found.
type userExternalAccountNotFoundError struct {
	args []interface{}
}

func (err userExternalAccountNotFoundError) Error() string {
	return fmt.Sprintf("user external account not found: %v", err.args)
}

func (err userExternalAccountNotFoundError) NotFound() bool {
	return true
}

// userExternalAccounts provides access to the `user_external_accounts` table.
type userExternalAccounts struct{}

// Get gets information about the user external account.
func (s *userExternalAccounts) Get(ctx context.Context, id int32) (*extsvc.ExternalAccount, error) {
	if Mocks.ExternalAccounts.Get != nil {
		return Mocks.ExternalAccounts.Get(id)
	}
	return s.getBySQL(ctx, sqlf.Sprintf("WHERE id=%d AND deleted_at IS NULL LIMIT 1", id))
}

// LookupUserAndSave is used for authenticating a user (when both their Sourcegraph account and the
// association with the external account already exist).
//
// It looks up the existing user associated with the external account's ExternalAccountSpec. If
// found, it updates the account's data and returns the user. It NEVER creates a user; you must call
// CreateUserAndSave for that.
func (s *userExternalAccounts) LookupUserAndSave(ctx context.Context, spec extsvc.ExternalAccountSpec, data extsvc.ExternalAccountData) (userID int32, err error) {
	if Mocks.ExternalAccounts.LookupUserAndSave != nil {
		return Mocks.ExternalAccounts.LookupUserAndSave(spec, data)
	}

	err = dbconn.Global.QueryRowContext(ctx, `
UPDATE user_external_accounts SET auth_data=$5, account_data=$6, updated_at=now()
WHERE service_type=$1 AND service_id=$2 AND client_id=$3 AND account_id=$4 AND deleted_at IS NULL
RETURNING user_id
`, spec.ServiceType, spec.ServiceID, spec.ClientID, spec.AccountID, data.AuthData, data.AccountData).Scan(&userID)
	if err == sql.ErrNoRows {
		err = userExternalAccountNotFoundError{[]interface{}{spec}}
	}
	return userID, err
}

// AssociateUserAndSave is used for linking a new, additional external account with an existing
// Sourcegraph account.
//
// It creates a user external account and associates it with the specified user. If the external
// account already exists and is associated with:
//
// - the same user: it updates the data and returns a nil error; or
// - a different user: it performs no update and returns a non-nil error
func (s *userExternalAccounts) AssociateUserAndSave(ctx context.Context, userID int32, spec extsvc.ExternalAccountSpec, data extsvc.ExternalAccountData) (err error) {
	if Mocks.ExternalAccounts.AssociateUserAndSave != nil {
		return Mocks.ExternalAccounts.AssociateUserAndSave(userID, spec, data)
	}

	// This "upsert" may cause us to return an ephemeral failure due to a race condition, but it
	// won't result in inconsistent data.  Wrap in transaction.
	tx, err := dbconn.Global.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				err = multierror.Append(err, rollErr)
			}
			return
		}
		err = tx.Commit()
	}()

	// Find whether the account exists and, if so, which user ID the account is associated with.
	var exists bool
	var existingID, associatedUserID int32
	err = tx.QueryRowContext(ctx, `
SELECT id, user_id FROM user_external_accounts
WHERE service_type=$1 AND service_id=$2 AND client_id=$3 AND account_id=$4 AND deleted_at IS NULL
`, spec.ServiceType, spec.ServiceID, spec.ClientID, spec.AccountID).Scan(&existingID, &associatedUserID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	exists = err != sql.ErrNoRows
	err = nil

	if exists && associatedUserID != userID {
		// The account already exists and is associated with another user.
		return fmt.Errorf("unable to change association of external account from user %d to user %d (delete the external account and then try again)", associatedUserID, userID)
	}

	if !exists {
		// Create the external account (it doesn't yet exist).
		return s.insert(ctx, tx, userID, spec, data)
	}

	// Update the external account (it exists).
	res, err := tx.ExecContext(ctx, `
UPDATE user_external_accounts SET auth_data=$6, account_data=$7, updated_at=now()
WHERE service_type=$1 AND service_id=$2 AND client_id=$3 AND account_id=$4 AND user_id=$5 AND deleted_at IS NULL
`, spec.ServiceType, spec.ServiceID, spec.ClientID, spec.AccountID, userID, data.AuthData, data.AccountData)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return userExternalAccountNotFoundError{[]interface{}{existingID}}
	}
	return nil
}

// CreateUserAndSave is used to create a new Sourcegraph user account from an external account
// (e.g., "signup from SAML").
//
// It creates a new user and associates it with the specified external account. If the user to
// create already exists, it returns an error.
func (s *userExternalAccounts) CreateUserAndSave(ctx context.Context, newUser NewUser, spec extsvc.ExternalAccountSpec, data extsvc.ExternalAccountData) (createdUserID int32, err error) {
	if Mocks.ExternalAccounts.CreateUserAndSave != nil {
		return Mocks.ExternalAccounts.CreateUserAndSave(newUser, spec, data)
	}

	// Wrap in transaction.
	tx, err := dbconn.Global.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				err = multierror.Append(err, rollErr)
			}
			return
		}
		err = tx.Commit()
	}()

	createdUser, err := Users.create(ctx, tx, newUser)
	if err != nil {
		return 0, err
	}

	err = s.insert(ctx, tx, createdUser.ID, spec, data)
	return createdUser.ID, err
}

func (s *userExternalAccounts) insert(ctx context.Context, tx *sql.Tx, userID int32, spec extsvc.ExternalAccountSpec, data extsvc.ExternalAccountData) error {
	_, err := tx.ExecContext(ctx, `
INSERT INTO user_external_accounts(user_id, service_type, service_id, client_id, account_id, auth_data, account_data)
VALUES($1, $2, $3, $4, $5, $6, $7)
`, userID, spec.ServiceType, spec.ServiceID, spec.ClientID, spec.AccountID, data.AuthData, data.AccountData)
	return err
}

// Delete deletes a user external account.
func (*userExternalAccounts) Delete(ctx context.Context, id int32) error {
	if Mocks.ExternalAccounts.Delete != nil {
		return Mocks.ExternalAccounts.Delete(id)
	}

	res, err := dbconn.Global.ExecContext(ctx, "UPDATE user_external_accounts SET deleted_at=now() WHERE id=$1 AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return userExternalAccountNotFoundError{[]interface{}{id}}
	}
	return nil
}

// ExternalAccountsListOptions specifies the options for listing user external accounts.
type ExternalAccountsListOptions struct {
	UserID                           int32
	ServiceType, ServiceID, ClientID string
	*LimitOffset
}

func (s *userExternalAccounts) List(ctx context.Context, opt ExternalAccountsListOptions) (acct []*extsvc.ExternalAccount, err error) {
	tr, ctx := trace.New(ctx, "userExternalAccounts.List", "")
	defer func() {
		if err != nil {
			tr.SetError(err)
		}

		tr.LogFields(
			otlog.Object("opt", opt),
			otlog.Int("accounts.count", len(acct)),
		)

		tr.Finish()
	}()

	if Mocks.ExternalAccounts.List != nil {
		return Mocks.ExternalAccounts.List(opt)
	}

	conds := s.listSQL(opt)
	return s.listBySQL(ctx, sqlf.Sprintf("WHERE %s ORDER BY id ASC %s", sqlf.Join(conds, "AND"), opt.LimitOffset.SQL()))
}

func (s *userExternalAccounts) Count(ctx context.Context, opt ExternalAccountsListOptions) (int, error) {
	if Mocks.ExternalAccounts.Count != nil {
		return Mocks.ExternalAccounts.Count(opt)
	}

	conds := s.listSQL(opt)
	q := sqlf.Sprintf("SELECT COUNT(*) FROM user_external_accounts WHERE %s", sqlf.Join(conds, "AND"))
	var count int
	err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count)
	return count, err
}

// TmpMigrate implements the migration described in bg.MigrateExternalAccounts (which is the only
// func that should call this).
func (*userExternalAccounts) TmpMigrate(ctx context.Context, serviceType string) error {
	// TEMP: Delete all external accounts associated with deleted users. Due to a bug in this
	// migration code, it was possible for deleted users to be associated with non-deleted external
	// accounts. This caused unexpected behavior in the UI (although did not pose a security
	// threat). So, run this cleanup task upon each server startup.
	if err := (userExternalAccounts{}).deleteForDeletedUsers(ctx); err != nil {
		log15.Warn("Unable to clean up external user accounts.", "err", err)
	}

	const needsMigrationSentinel = "migration_in_progress"

	// Avoid running UPDATE (which takes a lock) if it's not needed. The UPDATE only needs to run
	// once ever, and we are guaranteed that the DB migration has run by the time we arrive here, so
	// this is safe and not racy.
	var needsMigration bool
	if err := dbconn.Global.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_external_accounts WHERE service_type=$1 AND deleted_at IS NULL)`, needsMigrationSentinel).Scan(&needsMigration); err != nil && err != sql.ErrNoRows {
		return err
	}
	if !needsMigration {
		return nil
	}

	var err error
	if serviceType == "" {
		_, err = dbconn.Global.ExecContext(ctx, `UPDATE user_external_accounts SET deleted_at=now(), service_type='not_configured_at_migration_time' WHERE service_type=$1`, needsMigrationSentinel)
	} else {
		_, err = dbconn.Global.ExecContext(ctx, `UPDATE user_external_accounts SET service_type=$2, account_id=SUBSTR(account_id, CHAR_LENGTH(service_id)+2) WHERE service_type=$1 AND service_id!='override'`, needsMigrationSentinel, serviceType)
		if err == nil {
			_, err = dbconn.Global.ExecContext(ctx, `UPDATE user_external_accounts SET service_type='override', service_id='' WHERE service_type=$1 AND service_id='override'`, needsMigrationSentinel)
		}
	}
	return err
}

func (userExternalAccounts) deleteForDeletedUsers(ctx context.Context) error {
	_, err := dbconn.Global.ExecContext(ctx, `UPDATE user_external_accounts SET deleted_at=now() FROM users WHERE user_external_accounts.user_id=users.id AND users.deleted_at IS NOT NULL AND user_external_accounts.deleted_at IS NULL`)
	return err
}

func (s *userExternalAccounts) getBySQL(ctx context.Context, querySuffix *sqlf.Query) (*extsvc.ExternalAccount, error) {
	results, err := s.listBySQL(ctx, querySuffix)
	if err != nil {
		return nil, err
	}
	if len(results) != 1 {
		return nil, userExternalAccountNotFoundError{querySuffix.Args()}
	}
	return results[0], nil
}

func (*userExternalAccounts) listBySQL(ctx context.Context, querySuffix *sqlf.Query) ([]*extsvc.ExternalAccount, error) {
	q := sqlf.Sprintf(`SELECT t.id, t.user_id, t.service_type, t.service_id, t.client_id, t.account_id, t.auth_data, t.account_data, t.created_at, t.updated_at FROM user_external_accounts t %s`, querySuffix)
	rows, err := dbconn.Global.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	var results []*extsvc.ExternalAccount
	defer rows.Close()
	for rows.Next() {
		var o extsvc.ExternalAccount
		if err := rows.Scan(&o.ID, &o.UserID, &o.ServiceType, &o.ServiceID, &o.ClientID, &o.AccountID, &o.AuthData, &o.AccountData, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, &o)
	}
	return results, rows.Err()
}

func (*userExternalAccounts) listSQL(opt ExternalAccountsListOptions) (conds []*sqlf.Query) {
	conds = []*sqlf.Query{sqlf.Sprintf("deleted_at IS NULL")}

	if opt.UserID != 0 {
		conds = append(conds, sqlf.Sprintf("user_id=%d", opt.UserID))
	}
	if opt.ServiceType != "" || opt.ServiceID != "" || opt.ClientID != "" {
		conds = append(conds, sqlf.Sprintf("(service_type=%s AND service_id=%s AND client_id=%s)", opt.ServiceType, opt.ServiceID, opt.ClientID))
	}
	return conds
}

// MockExternalAccounts mocks the Stores.ExternalAccounts DB store.
type MockExternalAccounts struct {
	Get                  func(id int32) (*extsvc.ExternalAccount, error)
	LookupUserAndSave    func(extsvc.ExternalAccountSpec, extsvc.ExternalAccountData) (userID int32, err error)
	AssociateUserAndSave func(userID int32, spec extsvc.ExternalAccountSpec, data extsvc.ExternalAccountData) error
	CreateUserAndSave    func(NewUser, extsvc.ExternalAccountSpec, extsvc.ExternalAccountData) (createdUserID int32, err error)
	Delete               func(id int32) error
	List                 func(ExternalAccountsListOptions) ([]*extsvc.ExternalAccount, error)
	Count                func(ExternalAccountsListOptions) (int, error)
}
