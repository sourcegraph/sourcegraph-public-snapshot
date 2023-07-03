package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/types"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// userExternalAccountNotFoundError is the error that is returned when a user external account is not found.
type userExternalAccountNotFoundError struct {
	args []any
}

func (err userExternalAccountNotFoundError) Error() string {
	return fmt.Sprintf("user external account not found: %v", err.args)
}

func (err userExternalAccountNotFoundError) NotFound() bool {
	return true
}

// UserExternalAccountsStore provides access to the `user_external_accounts` table.
type UserExternalAccountsStore interface {
	// AssociateUserAndSave is used for linking a new, additional external account with an existing
	// Sourcegraph account.
	//
	// It creates a user external account and associates it with the specified user. If the external
	// account already exists and is associated with:
	//
	// - the same user: it updates the data and returns a nil error; or
	// - a different user: it performs no update and returns a non-nil error
	AssociateUserAndSave(ctx context.Context, userID int32, spec extsvc.AccountSpec, data extsvc.AccountData) (err error)

	Count(ctx context.Context, opt ExternalAccountsListOptions) (int, error)

	// CreateUserAndSave is used to create a new Sourcegraph user account from an external account
	// (e.g., "signup from SAML").
	//
	// It creates a new user and associates it with the specified external account. If the user to
	// create already exists, it returns an error.
	CreateUserAndSave(ctx context.Context, newUser NewUser, spec extsvc.AccountSpec, data extsvc.AccountData) (createdUser *types.User, err error)

	// Delete will soft delete all accounts matching the options combined using AND.
	// If options are all zero values then it does nothing.
	Delete(ctx context.Context, opt ExternalAccountsDeleteOptions) error

	// ExecResult performs a query without returning any rows, but includes the
	// result of the execution.
	ExecResult(ctx context.Context, query *sqlf.Query) (sql.Result, error)

	// Get gets information about the user external account.
	Get(ctx context.Context, id int32) (*extsvc.Account, error)

	// Insert creates the external account record in the database
	Insert(ctx context.Context, userID int32, spec extsvc.AccountSpec, data extsvc.AccountData) error

	List(ctx context.Context, opt ExternalAccountsListOptions) (acct []*extsvc.Account, err error)

	ListForUsers(ctx context.Context, userIDs []int32) (userToAccts map[int32][]*extsvc.Account, err error)

	// LookupUserAndSave is used for authenticating a user (when both their Sourcegraph account and the
	// association with the external account already exist).
	//
	// It looks up the existing user associated with the external account's extsvc.AccountSpec. If
	// found, it updates the account's data and returns the user. It NEVER creates a user; you must call
	// CreateUserAndSave for that.
	LookupUserAndSave(ctx context.Context, spec extsvc.AccountSpec, data extsvc.AccountData) (userID int32, err error)

	// UpsertSCIMData updates the external account data for the given user's SCIM account.
	// It looks up the existing user based on its ID, then sets its account ID and data.
	// Account ID is the same as the external ID for SCIM.
	// If the external account does not exist, it creates a new one.
	UpsertSCIMData(ctx context.Context, userID int32, accountID string, data extsvc.AccountData) (err error)

	// TouchExpired sets the given user external accounts to be expired now.
	TouchExpired(ctx context.Context, ids ...int32) error

	// TouchLastValid sets last valid time of the given user external account to be now.
	TouchLastValid(ctx context.Context, id int32) error

	WithEncryptionKey(key encryption.Key) UserExternalAccountsStore

	QueryRow(ctx context.Context, query *sqlf.Query) *sql.Row
	Transact(ctx context.Context) (UserExternalAccountsStore, error)
	With(other basestore.ShareableStore) UserExternalAccountsStore
	Done(error) error
	basestore.ShareableStore
}

type userExternalAccountsStore struct {
	*basestore.Store

	key encryption.Key

	logger log.Logger
}

// ExternalAccountsWith instantiates and returns a new UserExternalAccountsStore using the other store handle.
func ExternalAccountsWith(logger log.Logger, other basestore.ShareableStore) UserExternalAccountsStore {
	return &userExternalAccountsStore{logger: logger, Store: basestore.NewWithHandle(other.Handle())}
}

func (s *userExternalAccountsStore) With(other basestore.ShareableStore) UserExternalAccountsStore {
	return &userExternalAccountsStore{logger: s.logger, Store: s.Store.With(other), key: s.key}
}

func (s *userExternalAccountsStore) WithEncryptionKey(key encryption.Key) UserExternalAccountsStore {
	return &userExternalAccountsStore{logger: s.logger, Store: s.Store, key: key}
}

func (s *userExternalAccountsStore) Transact(ctx context.Context) (UserExternalAccountsStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &userExternalAccountsStore{logger: s.logger, Store: txBase, key: s.key}, err
}

func (s *userExternalAccountsStore) getEncryptionKey() encryption.Key {
	if s.key != nil {
		return s.key
	}
	return keyring.Default().UserExternalAccountKey
}

func (s *userExternalAccountsStore) Get(ctx context.Context, id int32) (*extsvc.Account, error) {
	return s.getBySQL(ctx, sqlf.Sprintf("WHERE id=%d AND deleted_at IS NULL LIMIT 1", id))
}

func (s *userExternalAccountsStore) LookupUserAndSave(ctx context.Context, spec extsvc.AccountSpec, data extsvc.AccountData) (userID int32, err error) {
	encryptedAuthData, encryptedAccountData, keyID, err := s.encryptData(ctx, data)
	if err != nil {
		return 0, err
	}

	err = s.Handle().QueryRowContext(ctx, `
UPDATE user_external_accounts
SET
	auth_data = $5,
	account_data = $6,
	encryption_key_id = $7,
	updated_at = now(),
	expired_at = NULL
WHERE
	service_type = $1
AND service_id = $2
AND client_id = $3
AND account_id = $4
AND deleted_at IS NULL
RETURNING user_id
`, spec.ServiceType, spec.ServiceID, spec.ClientID, spec.AccountID, encryptedAuthData, encryptedAccountData, keyID).Scan(&userID)
	if err == sql.ErrNoRows {
		err = userExternalAccountNotFoundError{[]any{spec}}
	}
	return userID, err
}

func (s *userExternalAccountsStore) UpsertSCIMData(ctx context.Context, userID int32, accountID string, data extsvc.AccountData) (err error) {
	encryptedAuthData, encryptedAccountData, keyID, err := s.encryptData(ctx, data)
	if err != nil {
		return
	}

	res, err := s.ExecResult(ctx, sqlf.Sprintf(`
UPDATE user_external_accounts
SET
	account_id = %s,
	auth_data = %s,
	account_data = %s,
	encryption_key_id = %s,
	updated_at = now(),
	expired_at = NULL
WHERE
	user_id = %s
AND service_type = %s
AND service_id = %s
AND deleted_at IS NULL
`, accountID, encryptedAuthData, encryptedAccountData, keyID, userID, "scim", "scim"))
	if err != nil {
		return
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return
	}

	if rowsAffected == 0 {
		return s.Insert(ctx, userID, extsvc.AccountSpec{ServiceType: "scim", ServiceID: "scim", AccountID: accountID}, data)
	}

	// This logs an audit event for account changes but only if they are initiated via SCIM
	logAccountModifiedEvent(ctx, NewDBWith(s.logger, s), userID, "scim")

	return
}

func (s *userExternalAccountsStore) AssociateUserAndSave(ctx context.Context, userID int32, spec extsvc.AccountSpec, data extsvc.AccountData) (err error) {
	// This "upsert" may cause us to return an ephemeral failure due to a race condition, but it
	// won't result in inconsistent data.  Wrap in transaction.

	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Find whether the account exists and, if so, which user ID the account is associated with.
	var exists bool
	var existingID, associatedUserID int32
	err = tx.QueryRow(ctx, sqlf.Sprintf(`
SELECT id, user_id
FROM user_external_accounts
WHERE
	service_type = %s
AND service_id = %s
AND client_id = %s
AND account_id = %s
AND deleted_at IS NULL
`, spec.ServiceType, spec.ServiceID, spec.ClientID, spec.AccountID)).Scan(&existingID, &associatedUserID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	exists = err != sql.ErrNoRows
	err = nil

	if exists && associatedUserID != userID {
		// The account already exists and is associated with another user.
		return errors.Errorf("unable to change association of external account from user %d to user %d (delete the external account and then try again)", associatedUserID, userID)
	}

	if !exists {
		// Create the external account (it doesn't yet exist).
		return tx.Insert(ctx, userID, spec, data)
	}

	var encryptedAuthData, encryptedAccountData, keyID string
	if data.AuthData != nil {
		encryptedAuthData, keyID, err = data.AuthData.Encrypt(ctx, s.getEncryptionKey())
		if err != nil {
			return err
		}
	}
	if data.Data != nil {
		encryptedAccountData, keyID, err = data.Data.Encrypt(ctx, s.getEncryptionKey())
		if err != nil {
			return err
		}
	}

	// Update the external account (it exists).
	res, err := tx.ExecResult(ctx, sqlf.Sprintf(`
UPDATE user_external_accounts
SET
	auth_data = %s,
	account_data = %s,
	encryption_key_id = %s,
	updated_at = now(),
	expired_at = NULL
WHERE
	service_type = %s
AND service_id = %s
AND client_id = %s
AND account_id = %s
AND user_id = %s
AND deleted_at IS NULL
`, encryptedAuthData, encryptedAccountData, keyID, spec.ServiceType, spec.ServiceID, spec.ClientID, spec.AccountID, userID))
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return userExternalAccountNotFoundError{[]any{existingID}}
	}
	return nil
}

func (s *userExternalAccountsStore) CreateUserAndSave(ctx context.Context, newUser NewUser, spec extsvc.AccountSpec, data extsvc.AccountData) (createdUser *types.User, err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	createdUser, err = UsersWith(s.logger, tx).CreateInTransaction(ctx, newUser, &spec)
	if err != nil {
		return nil, err
	}

	err = tx.Insert(ctx, createdUser.ID, spec, data)
	if err == nil {
		logAccountCreatedEvent(ctx, NewDBWith(s.logger, s), createdUser, spec.ServiceType)
	}
	return createdUser, err
}

func (s *userExternalAccountsStore) Insert(ctx context.Context, userID int32, spec extsvc.AccountSpec, data extsvc.AccountData) (err error) {
	encryptedAuthData, encryptedAccountData, keyID, err := s.encryptData(ctx, data)
	if err != nil {
		return
	}

	return s.Exec(ctx, sqlf.Sprintf(`
INSERT INTO user_external_accounts (user_id, service_type, service_id, client_id, account_id, auth_data, account_data, encryption_key_id)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
`, userID, spec.ServiceType, spec.ServiceID, spec.ClientID, spec.AccountID, encryptedAuthData, encryptedAccountData, keyID))
}

// encryptData encrypts the given account data and returns the encrypted data and key ID.
func (s *userExternalAccountsStore) encryptData(ctx context.Context, accountData extsvc.AccountData) (eAuthData string, eData string, keyID string, err error) {
	if accountData.AuthData != nil {
		eAuthData, keyID, err = accountData.AuthData.Encrypt(ctx, s.getEncryptionKey())
		if err != nil {
			return
		}
	}
	if accountData.Data != nil {
		eData, keyID, err = accountData.Data.Encrypt(ctx, s.getEncryptionKey())
	}
	return
}

func (s *userExternalAccountsStore) TouchExpired(ctx context.Context, ids ...int32) error {
	if len(ids) == 0 {
		return nil
	}

	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = strconv.Itoa(int(id))
	}
	_, err := s.Handle().ExecContext(ctx, fmt.Sprintf(`
UPDATE user_external_accounts
SET expired_at = now()
WHERE id IN (%s)
`, strings.Join(idStrings, ", ")))
	return err
}

func (s *userExternalAccountsStore) TouchLastValid(ctx context.Context, id int32) error {
	_, err := s.Handle().ExecContext(ctx, `
UPDATE user_external_accounts
SET
	expired_at = NULL,
	last_valid_at = now()
WHERE id = $1
`, id)
	return err
}

// ExternalAccountsDeleteOptions defines criteria that will be used to select
// which accounts to soft delete.
type ExternalAccountsDeleteOptions struct {
	// A slice of ExternalAccountIDs
	IDs         []int32
	UserID      int32
	AccountID   string
	ServiceType string
}

// Delete will soft delete all accounts matching the options combined using AND.
// If options are all zero values then it does nothing.
func (s *userExternalAccountsStore) Delete(ctx context.Context, opt ExternalAccountsDeleteOptions) error {
	conds := []*sqlf.Query{sqlf.Sprintf("deleted_at IS NULL")}

	if len(opt.IDs) > 0 {
		ids := make([]*sqlf.Query, len(opt.IDs))
		for i, id := range opt.IDs {
			ids[i] = sqlf.Sprintf("%s", id)
		}
		conds = append(conds, sqlf.Sprintf("id IN (%s)", sqlf.Join(ids, ",")))
	}
	if opt.UserID != 0 {
		conds = append(conds, sqlf.Sprintf("user_id=%d", opt.UserID))
	}
	if opt.AccountID != "" {
		conds = append(conds, sqlf.Sprintf("account_id=%s", opt.AccountID))
	}
	if opt.ServiceType != "" {
		conds = append(conds, sqlf.Sprintf("service_type=%s", opt.ServiceType))
	}

	// We only have the default deleted_at clause, do nothing
	if len(conds) == 1 {
		return nil
	}

	q := sqlf.Sprintf(`
UPDATE user_external_accounts
SET deleted_at=now()
WHERE %s`, sqlf.Join(conds, "AND"))

	err := s.Exec(ctx, q)

	return errors.Wrap(err, "executing delete")
}

// ExternalAccountsListOptions specifies the options for listing user external accounts.
type ExternalAccountsListOptions struct {
	UserID      int32
	ServiceType string
	ServiceID   string
	ClientID    string
	AccountID   string

	// Only one of these should be set
	ExcludeExpired bool
	OnlyExpired    bool

	*LimitOffset
}

func (s *userExternalAccountsStore) List(ctx context.Context, opt ExternalAccountsListOptions) (acct []*extsvc.Account, err error) {
	tr, ctx := trace.DeprecatedNew(ctx, "UserExternalAccountsStore.List", "")
	defer func() {
		if err != nil {
			tr.SetError(err)
		}

		tr.AddEvent(
			"done",
			attribute.String("opt", fmt.Sprintf("%#v", opt)),
			attribute.Int("accounts.count", len(acct)),
		)

		tr.Finish()
	}()

	conds := s.listSQL(opt)
	return s.listBySQL(ctx, sqlf.Sprintf("WHERE %s ORDER BY id ASC %s", sqlf.Join(conds, "AND"), opt.LimitOffset.SQL()))
}

func (s *userExternalAccountsStore) ListForUsers(ctx context.Context, userIDs []int32) (userToAccts map[int32][]*extsvc.Account, err error) {
	tr, ctx := trace.DeprecatedNew(ctx, "UserExternalAccountsStore.ListForUsers", "")
	var count int
	defer func() {
		if err != nil {
			tr.SetError(err)
		}
		tr.AddEvent(
			"done",
			attribute.String("userIDs", fmt.Sprintf("%v", userIDs)),
			attribute.Int("accounts.count", count),
		)
		tr.Finish()
	}()
	if len(userIDs) == 0 {
		return
	}
	condition := sqlf.Sprintf("WHERE user_id = ANY(%s)", pq.Array(userIDs))
	accts, err := s.listBySQL(ctx, condition)
	if err != nil {
		return nil, err
	}
	count = len(accts)
	userToAccts = make(map[int32][]*extsvc.Account)
	for _, acct := range accts {
		userID := acct.UserID
		if _, ok := userToAccts[userID]; !ok {
			userToAccts[userID] = make([]*extsvc.Account, 0)
		}
		userToAccts[userID] = append(userToAccts[userID], acct)
	}
	return
}

func (s *userExternalAccountsStore) Count(ctx context.Context, opt ExternalAccountsListOptions) (int, error) {
	conds := s.listSQL(opt)
	q := sqlf.Sprintf("SELECT COUNT(*) FROM user_external_accounts WHERE %s", sqlf.Join(conds, "AND"))
	var count int
	err := s.QueryRow(ctx, q).Scan(&count)
	return count, err
}

func (s *userExternalAccountsStore) getBySQL(ctx context.Context, querySuffix *sqlf.Query) (*extsvc.Account, error) {
	results, err := s.listBySQL(ctx, querySuffix)
	if err != nil {
		return nil, err
	}
	if len(results) != 1 {
		return nil, userExternalAccountNotFoundError{querySuffix.Args()}
	}
	return results[0], nil
}

func (s *userExternalAccountsStore) listBySQL(ctx context.Context, querySuffix *sqlf.Query) ([]*extsvc.Account, error) {
	q := sqlf.Sprintf(`
SELECT
    t.id,
    t.user_id,
    t.service_type,
    t.service_id,
    t.client_id,
    t.account_id,
    t.auth_data,
    t.account_data,
    t.created_at,
    t.updated_at,
    t.encryption_key_id
FROM user_external_accounts t
%s`, querySuffix)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*extsvc.Account
	for rows.Next() {
		var acct extsvc.Account
		var authData, accountData sql.NullString
		var keyID string
		if err := rows.Scan(
			&acct.ID, &acct.UserID,
			&acct.ServiceType, &acct.ServiceID, &acct.ClientID, &acct.AccountID,
			&authData, &accountData,
			&acct.CreatedAt, &acct.UpdatedAt,
			&keyID,
		); err != nil {
			return nil, err
		}

		if authData.Valid {
			acct.AuthData = extsvc.NewEncryptedData(authData.String, keyID, s.getEncryptionKey())
		}
		if accountData.Valid {
			acct.Data = extsvc.NewEncryptedData(accountData.String, keyID, s.getEncryptionKey())
		}

		results = append(results, &acct)
	}
	return results, rows.Err()
}

func (s *userExternalAccountsStore) listSQL(opt ExternalAccountsListOptions) (conds []*sqlf.Query) {
	conds = []*sqlf.Query{sqlf.Sprintf("deleted_at IS NULL")}

	if opt.UserID != 0 {
		conds = append(conds, sqlf.Sprintf("user_id=%d", opt.UserID))
	}
	if opt.ServiceType != "" {
		conds = append(conds, sqlf.Sprintf("service_type=%s", opt.ServiceType))
	}
	if opt.ServiceID != "" {
		conds = append(conds, sqlf.Sprintf("service_id=%s", opt.ServiceID))
	}
	if opt.ClientID != "" {
		conds = append(conds, sqlf.Sprintf("client_id=%s", opt.ClientID))
	}
	if opt.AccountID != "" {
		conds = append(conds, sqlf.Sprintf("account_id=%s", opt.AccountID))
	}
	if opt.ExcludeExpired {
		conds = append(conds, sqlf.Sprintf("expired_at IS NULL"))
	}
	if opt.OnlyExpired {
		conds = append(conds, sqlf.Sprintf("expired_at IS NOT NULL"))
	}

	return conds
}
