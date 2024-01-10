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

	"github.com/sourcegraph/sourcegraph/internal/actor"
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
	// Upsert either creates or updates a user external account.
	// acct.UserID should match an existing user ID.
	//
	// If an external account with the same AccountSpec already exists, the
	// user ID associated with the existing account must match acct.UserID,
	// otherwise no update will be performed and an error will be returned.
	Upsert(ctx context.Context, acct *extsvc.Account) (*extsvc.Account, error)

	Count(ctx context.Context, opt ExternalAccountsListOptions) (int, error)

	// Delete will soft (or hard) delete all accounts matching the options combined using AND.
	// If options are all zero values then it does nothing.
	Delete(ctx context.Context, opt ExternalAccountsDeleteOptions) error

	// ExecResult performs a query without returning any rows, but includes the
	// result of the execution.
	ExecResult(ctx context.Context, query *sqlf.Query) (sql.Result, error)

	// Get gets information about the user external account.
	Get(ctx context.Context, id int32) (*extsvc.Account, error)

	// Insert creates the external account record in the database and returns it.
	Insert(ctx context.Context, acct *extsvc.Account) (*extsvc.Account, error)

	List(ctx context.Context, opt ExternalAccountsListOptions) (acct []*extsvc.Account, err error)

	ListForUsers(ctx context.Context, userIDs []int32) (userToAccts map[int32][]*extsvc.Account, err error)

	// Update updates an existing external account in the database that matches acct.AccountSpec.
	// The updated external account is returned, and will contain fields from the database
	// such as the corresponding user ID.
	Update(ctx context.Context, acct *extsvc.Account) (*extsvc.Account, error)

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

func (s *userExternalAccountsStore) Update(ctx context.Context, acct *extsvc.Account) (*extsvc.Account, error) {
	encryptedAuthData, encryptedAccountData, keyID, err := s.encryptData(ctx, acct.AccountData)
	if err != nil {
		return nil, err
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
RETURNING id, user_id, updated_at, created_at
`, acct.AccountSpec.ServiceType, acct.AccountSpec.ServiceID, acct.AccountSpec.ClientID, acct.AccountSpec.AccountID, encryptedAuthData, encryptedAccountData, keyID).Scan(&acct.ID, &acct.UserID, &acct.UpdatedAt, &acct.CreatedAt)
	if err == sql.ErrNoRows {
		err = userExternalAccountNotFoundError{[]any{acct.AccountSpec}}
	}
	return acct, err
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
		_, err := s.Insert(ctx,
			&extsvc.Account{
				UserID:      userID,
				AccountSpec: extsvc.AccountSpec{ServiceType: "scim", ServiceID: "scim", AccountID: accountID},
				AccountData: data,
			})
		return err
	}

	// This logs an audit event for account changes but only if they are initiated via SCIM
	arg := struct {
		Modifier    int32  `json:"modifier"`
		ServiceType string `json:"service_type"`
	}{
		Modifier:    actor.FromContext(ctx).UID,
		ServiceType: "scim",
	}
	if err := NewDBWith(s.logger, s).SecurityEventLogs().LogSecurityEvent(ctx, SecurityEventNameAccountModified, "", uint32(userID), "", "scim", arg); err != nil {
		s.logger.Warn("Error logging security event", log.Error(err))
	}

	return
}

func (s *userExternalAccountsStore) Upsert(ctx context.Context, acct *extsvc.Account) (_ *extsvc.Account, err error) {
	// This "upsert" may cause us to return an ephemeral failure due to a race condition, but it
	// won't result in inconsistent data.  Wrap in transaction.

	tx, err := s.Transact(ctx)
	if err != nil {
		return nil, err
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
`, acct.AccountSpec.ServiceType, acct.AccountSpec.ServiceID, acct.AccountSpec.ClientID, acct.AccountSpec.AccountID)).Scan(&existingID, &associatedUserID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	exists = err != sql.ErrNoRows
	err = nil

	if exists && associatedUserID != acct.UserID {
		// The account already exists and is associated with another user.
		return nil, errors.Errorf("unable to change association of external account from user %d to user %d (delete the external account and then try again)", associatedUserID, acct.UserID)
	}

	if !exists {
		// Create the external account (it doesn't yet exist).
		return tx.Insert(ctx, acct)
	}

	var encryptedAuthData, encryptedAccountData, keyID string
	if acct.AccountData.AuthData != nil {
		encryptedAuthData, keyID, err = acct.AccountData.AuthData.Encrypt(ctx, s.getEncryptionKey())
		if err != nil {
			return nil, err
		}
	}
	if acct.AccountData.Data != nil {
		encryptedAccountData, keyID, err = acct.AccountData.Data.Encrypt(ctx, s.getEncryptionKey())
		if err != nil {
			return nil, err
		}
	}

	// Update the external account (it exists).
	res := tx.QueryRow(ctx, sqlf.Sprintf(`
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
RETURNING
	id, updated_at, created_at
`, encryptedAuthData, encryptedAccountData, keyID, acct.AccountSpec.ServiceType, acct.AccountSpec.ServiceID, acct.AccountSpec.ClientID, acct.AccountSpec.AccountID, acct.UserID))
	if res.Err() != nil {
		return nil, res.Err()
	}

	if err := res.Scan(&acct.ID, &acct.UpdatedAt, &acct.CreatedAt); err != nil {
		return nil, err
	}
	return acct, nil
}

func (s *userExternalAccountsStore) Insert(ctx context.Context, acct *extsvc.Account) (_ *extsvc.Account, err error) {
	encryptedAuthData, encryptedAccountData, keyID, err := s.encryptData(ctx, acct.AccountData)
	if err != nil {
		return
	}

	res := s.QueryRow(ctx, sqlf.Sprintf(`
INSERT INTO user_external_accounts (user_id, service_type, service_id, client_id, account_id, auth_data, account_data, encryption_key_id)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
RETURNING id, updated_at, created_at
`, acct.UserID, acct.AccountSpec.ServiceType, acct.AccountSpec.ServiceID, acct.AccountSpec.ClientID, acct.AccountSpec.AccountID, encryptedAuthData, encryptedAccountData, keyID))

	err = res.Err()
	if err != nil {
		return nil, err
	}
	if err := res.Scan(&acct.ID, &acct.UpdatedAt, &acct.CreatedAt); err != nil {
		return nil, err
	}

	return acct, nil
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
// which accounts to soft (or hard) delete
type ExternalAccountsDeleteOptions struct {
	// A slice of ExternalAccountIDs
	IDs         []int32
	UserID      int32
	AccountID   string
	ServiceType string
	// HardDelete completely deletes any matching external accounts.
	HardDelete bool
}

// Delete will soft (or hard) delete all accounts matching the options combined using AND.
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

	var q *sqlf.Query
	if opt.HardDelete {
		q = sqlf.Sprintf(`
		DELETE FROM user_external_accounts
		WHERE %s`, sqlf.Join(conds, "AND"))
	} else {
		q = sqlf.Sprintf(`
		UPDATE user_external_accounts
		SET deleted_at=now()
		WHERE %s`, sqlf.Join(conds, "AND"))
	}

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
	tr, ctx := trace.New(ctx, "UserExternalAccountsStore.List")
	defer func() {
		if err != nil {
			tr.SetError(err)
		}

		tr.AddEvent(
			"done",
			attribute.String("opt", fmt.Sprintf("%#v", opt)),
			attribute.Int("accounts.count", len(acct)),
		)

		tr.End()
	}()

	conds := s.listSQL(opt)
	return s.listBySQL(ctx, sqlf.Sprintf("WHERE %s ORDER BY id ASC %s", sqlf.Join(conds, "AND"), opt.LimitOffset.SQL()))
}

func (s *userExternalAccountsStore) ListForUsers(ctx context.Context, userIDs []int32) (userToAccts map[int32][]*extsvc.Account, err error) {
	tr, ctx := trace.New(ctx, "UserExternalAccountsStore.ListForUsers")
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
		tr.End()
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
