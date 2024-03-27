package database

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/accesstoken"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/hashutil"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// AccessToken describes an access token. The actual token (that a caller must supply to
// authenticate) is not stored and is not present in this struct.
type AccessToken struct {
	ID            int64
	SubjectUserID int32 // the user whose privileges the access token grants
	Scopes        []string
	Note          string
	CreatorUserID int32
	// Internal determines whether or not the token shows up in the UI. Tokens
	// with internal=true are to be used with the executor service.
	Internal   bool
	CreatedAt  time.Time
	LastUsedAt *time.Time
	// ExpiresAt denotes the server time after which the token shall no longer be
	// valid and not grant access anymore.
	// IsZero will be true for access tokens without expiry.
	ExpiresAt time.Time
}

// ErrAccessTokenNotFound occurs when a database operation expects a specific access token to exist
// but it does not exist.
var ErrAccessTokenNotFound = errors.New("access token not found")

// ErrTooManyAccessTokens is returned when creating an access token would exceed the configured maximum number of active access tokens.
var ErrTooManyAccessTokens = errors.New("number of active access tokens exceeds limit")

// MaxAccessTokenLastUsedAtAge is the maximum amount of time we will wait before updating an access token's
// last_used_at column. We are OK letting the value get a little stale in order to cut down on database writes.
const MaxAccessTokenLastUsedAtAge = 5 * time.Minute

// InvalidTokenError is returned when decoding the hex encoded token passed to any
// of the methods on AccessTokenStore.
type InvalidTokenError struct {
	err error
}

func (e InvalidTokenError) Error() string {
	return fmt.Sprintf("invalid token: %s", e.err)
}

// AccessTokenStore implements autocert.Cache
type AccessTokenStore interface {
	// Count counts all access tokens, except internal tokens, that satisfy the options (ignoring limit and offset).
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the tokens.
	Count(context.Context, AccessTokensListOptions) (int, error)

	// Create creates an access token for the specified user. The secret token value itself is
	// returned. The caller is responsible for presenting this value to the end user; Sourcegraph does
	// not retain it (only a hash of it).
	//
	// The secret token value consists of the prefix "sgp_" and then a long random string. It is
	// what API clients must provide to authenticate their requests. We store the SHA-256 hash of
	// the secret token value in the database. This lets us verify a token's validity (in the
	// (*accessTokens).Lookup method) quickly, while still ensuring that an attacker who obtains the
	// access_tokens DB table would not be able to impersonate a token holder. We don't use bcrypt
	// because the original secret is a randomly generated string (not a password), so it's
	// implausible for an attacker to brute-force the input space; also bcrypt is slow and would add
	// noticeable latency to each request that supplied a token.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to create tokens for the
	// specified user (i.e., that the actor is either the user or a site admin).
	Create(ctx context.Context, subjectUserID int32, scopes []string, note string, creatorUserID int32, expiresAt time.Time) (id int64, token string, err error)

	// CreateInternal creates an *internal* access token for the specified user. An
	// internal access token will be used by Sourcegraph to talk to its API from
	// other services, i.e. executor jobs. Internal tokens do not show up in the UI.
	//
	// See the documentation for Create for more details.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to create tokens for the
	// specified user (i.e., that the actor is either the user or a site admin).
	CreateInternal(ctx context.Context, subjectUserID int32, scopes []string, note string, creatorUserID int32) (id int64, token string, err error)

	// GetOrCreateInternalToken returns the SHA256 hash of a random internal access token for the
	// given user. If no internal token exists, it creates one.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to view access tokens and
	// create tokens for the specified user (i.e., that the actor is either the user or a site admin).
	GetOrCreateInternalToken(ctx context.Context, subjectUserID int32, scopes []string) ([]byte, error)

	// DeleteByID deletes an access token given its ID.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to delete the token.
	DeleteByID(context.Context, int64) error

	// DeleteByToken deletes an access token given the secret token value itself (i.e., the same value
	// that an API client would use to authenticate). The token prefix "sgp_", if present, is stripped.
	DeleteByToken(ctx context.Context, token string) error

	// GetByID retrieves the access token (if any) given its ID.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this access token.
	GetByID(context.Context, int64) (*AccessToken, error)

	// GetByToken retrieves the access token (if any). The token prefix "sgp_", if present, is stripped.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this access token.
	GetByToken(ctx context.Context, token string) (*AccessToken, error)

	// HardDeleteByID hard-deletes an access token given its ID.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to delete the token.
	HardDeleteByID(context.Context, int64) error

	// List lists all access tokens that satisfy the options, except internal tokens.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
	// options.
	List(context.Context, AccessTokensListOptions) ([]*AccessToken, error)

	// Lookup looks up the access token. If it's valid and contains the required scope, it returns the
	// subject's user ID. Otherwise ErrAccessTokenNotFound is returned.
	//
	// The token prefix "sgp_", if present, is stripped.
	//
	// Calling Lookup also updates the access token's last-used-at date as applicable.
	//
	// 🚨 SECURITY: This returns a user ID if and only if the token corresponds to a valid,
	// non-deleted access token.
	Lookup(ctx context.Context, token string, opts TokenLookupOpts) (subjectUserID int32, err error)

	WithTransact(context.Context, func(AccessTokenStore) error) error
	With(basestore.ShareableStore) AccessTokenStore
	basestore.ShareableStore
}

type accessTokenStore struct {
	*basestore.Store
	logger log.Logger
}

var _ AccessTokenStore = (*accessTokenStore)(nil)

// AccessTokensWith instantiates and returns a new AccessTokenStore using the other store handle.
func AccessTokensWith(other basestore.ShareableStore, logger log.Logger) AccessTokenStore {
	return &accessTokenStore{Store: basestore.NewWithHandle(other.Handle()), logger: logger}
}

func (s *accessTokenStore) With(other basestore.ShareableStore) AccessTokenStore {
	return &accessTokenStore{Store: s.Store.With(other), logger: s.logger}
}

func (s *accessTokenStore) WithTransact(ctx context.Context, f func(AccessTokenStore) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&accessTokenStore{Store: tx, logger: s.logger})
	})
}

func (s *accessTokenStore) Create(ctx context.Context, subjectUserID int32, scopes []string, note string, creatorUserID int32, expiresAt time.Time) (id int64, token string, err error) {
	return s.createToken(ctx, subjectUserID, scopes, note, creatorUserID, expiresAt, false)
}

func (s *accessTokenStore) CreateInternal(ctx context.Context, subjectUserID int32, scopes []string, note string, creatorUserID int32) (id int64, token string, err error) {
	return s.createToken(ctx, subjectUserID, scopes, note, creatorUserID, time.Time{}, true)
}

func (s *accessTokenStore) createToken(ctx context.Context, subjectUserID int32, scopes []string, note string, creatorUserID int32, expiresAt time.Time, internal bool) (id int64, token string, err error) {
	if len(scopes) == 0 {
		// Prevent mistakes. There is no point in creating an access token with no scopes, and the
		// GraphQL API wouldn't let you do so anyway.
		return 0, "", errors.New("access tokens without scopes are not supported")
	}

	config := conf.Get().SiteConfig()

	var isDevInstance bool
	licenseInfo, err := licensing.GetConfiguredProductLicenseInfo()
	if err != nil || licenseInfo == nil {
		isDevInstance = true
	} else {
		isDevInstance = licensing.IsLicensePublicKeyOverridden()
	}

	token, b, err := accesstoken.GeneratePersonalAccessToken(config.LicenseKey, isDevInstance)
	if err != nil {
		return 0, "", err
	}

	if err := s.Handle().QueryRowContext(ctx,
		// Include users table query (with "FOR UPDATE") to ensure that subject/creator users have
		// not been deleted. If they were deleted, the query will return an error.
		`
WITH subject_user AS (
  SELECT
    id,
	(
		SELECT COUNT(*)
		FROM access_tokens
		WHERE
		  subject_user_id = $1
		  AND deleted_at IS NULL
		  AND (expires_at IS NULL OR expires_at > NOW())
		  AND internal IS NOT TRUE
	) AS active_tokens
   FROM users WHERE id=$1 AND deleted_at IS NULL FOR UPDATE
),
creator_user AS (
  SELECT id FROM users WHERE id=$5 AND deleted_at IS NULL FOR UPDATE
),
insert_values AS (
  SELECT subject_user.id AS subject_user_id, $2::text[] AS scopes, $3::bytea AS value_sha256, $4::text AS note, creator_user.id AS creator_user_id, $6::timestamp with time zone AS expires_at, $7::boolean AS internal
  FROM subject_user, creator_user
  WHERE subject_user.active_tokens < $8::int OR $7::boolean
)
INSERT INTO access_tokens(subject_user_id, scopes, value_sha256, note, creator_user_id, expires_at, internal) SELECT * FROM insert_values RETURNING id
`,
		subjectUserID, pq.Array(scopes), hashutil.ToSHA256Bytes(b[:]), note, creatorUserID, dbutil.NullTimeColumn(expiresAt), internal, conf.AccessTokensMaxPerUser(),
	).Scan(&id); err != nil {
		// if creation failed check to see if it was because too many tokens already
		count, countErr := s.Count(ctx, AccessTokensListOptions{SubjectUserID: subjectUserID})
		// if checking the count fails just return the original error
		if countErr != nil {
			return 0, "", err
		}
		if count >= conf.AccessTokensMaxPerUser() {
			return 0, "", ErrTooManyAccessTokens
		}
		return 0, "", err
	}

	// only log access tokens created by users
	if !internal {
		arg := struct {
			SubjectUserId int32     `json:"subject_user_id"`
			CreatorUserId int32     `json:"creator_user_id"`
			Scopes        []string  `json:"scopes"`
			Note          string    `json:"note"`
			ExpiresAt     time.Time `json:"expires_at"`
		}{
			SubjectUserId: subjectUserID,
			CreatorUserId: creatorUserID,
			Scopes:        scopes,
			Note:          note,
			ExpiresAt:     expiresAt,
		}
		if err := NewDBWith(s.logger, s).SecurityEventLogs().LogSecurityEvent(ctx, SecurityEventAccessTokenCreated, "", uint32(creatorUserID), "", "BACKEND", arg); err != nil {
			s.logger.Warn("Failed to log security event", log.Error(err))
		}
	}

	return id, token, nil
}

func (s *accessTokenStore) GetOrCreateInternalToken(ctx context.Context, subjectUserID int32, scopes []string) ([]byte, error) {
	sha256, err := s.getInternalToken(ctx, subjectUserID)
	if err != nil {
		_, _, err = s.CreateInternal(ctx, subjectUserID, scopes, "Created by GetOrCreateInternalToken", subjectUserID)
		if err != nil {
			return nil, err
		}
		return s.getInternalToken(ctx, subjectUserID)
	}
	return sha256, nil
}

// getInternalToken returns the SHA256 hash of a random internal access token for the given user.
func (s *accessTokenStore) getInternalToken(ctx context.Context, subjectUserID int32) ([]byte, error) {
	conds := []*sqlf.Query{
		sqlf.Sprintf("subject_user_id=%d", subjectUserID),
		sqlf.Sprintf("deleted_at IS NULL"),
		sqlf.Sprintf("internal IS TRUE"),
	}
	q := sqlf.Sprintf(`SELECT value_sha256 FROM access_tokens WHERE (%s) LIMIT 1`,
		sqlf.Join(conds, ") AND ("),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var sha256 []byte
	rows.Next()
	err = rows.Scan(&sha256)
	if err != nil {
		return nil, err
	}
	return sha256, nil
}

type TokenLookupOpts struct {
	OnlyAdmin     bool
	RequiredScope string
}

// Returns a query to upload the token's LastUsedAt column to "now". The returned query
// requires one parameter: the token ID to update.
func (o TokenLookupOpts) toUpdateLastUsedQuery() string {
	return `
	UPDATE access_tokens t SET last_used_at=now()
	WHERE t.id=$1 AND t.deleted_at IS NULL
`
}

// toGetQuery returns a SQL query that will return the following columns from
// the access_tokens table.
// - id
// - subject_user_id
// - last_used_at
//
// The query requires two parameters: the token's value_sha256 and scopes.
func (o TokenLookupOpts) toGetQuery() string {
	query := `
	SELECT t.id, t.subject_user_id, t.last_used_at
	FROM access_tokens t
	JOIN users subject_user ON t.subject_user_id=subject_user.id AND subject_user.deleted_at IS NULL
	JOIN users creator_user ON t.creator_user_id=creator_user.id AND creator_user.deleted_at IS NULL
	WHERE
	    t.value_sha256=$1
	    AND
		t.deleted_at IS NULL
		AND
		(t.expires_at IS NULL OR t.expires_at > NOW())
		AND
	    $2 = ANY (t.scopes)
`

	if o.OnlyAdmin {
		query += "AND subject_user.site_admin"
	}

	return query
}

func (s *accessTokenStore) Lookup(ctx context.Context, token string, opts TokenLookupOpts) (subjectUserID int32, err error) {
	if opts.RequiredScope == "" {
		return 0, errors.New("no scope provided in access token lookup")
	}

	tokenHash, err := tokenSHA256Hash(token)
	if err != nil {
		return 0, errors.Wrap(err, "AccessTokens.Lookup")
	}

	var (
		tokenID    int64
		subjectID  int32
		lastUsedAt *time.Time
	)
	row := s.Handle().QueryRowContext(ctx,
		// Ensure that subject and creator users still exist.
		opts.toGetQuery(),
		tokenHash, opts.RequiredScope)
	err = row.Scan(&tokenID, &subjectID, &lastUsedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrAccessTokenNotFound
		}
		return 0, err
	}

	if lastUsedAt == nil || time.Until(*lastUsedAt) < -MaxAccessTokenLastUsedAtAge {
		logger := s.logger.With(log.Int64("tokenID", tokenID), log.Int32("subjectID", subjectID))
		_, err := s.Handle().ExecContext(ctx, opts.toUpdateLastUsedQuery(), tokenID)
		if err != nil {
			logger.Warn("error trying to update token's last_used_at value", log.Error(err))
		} else {
			logger.Debug("updated access token's last_used_at value")
		}
	}

	return subjectID, nil
}

func (s *accessTokenStore) GetByID(ctx context.Context, id int64) (*AccessToken, error) {
	return s.get(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)})
}

func (s *accessTokenStore) GetByToken(ctx context.Context, token string) (*AccessToken, error) {
	tokenHash, err := tokenSHA256Hash(token)
	if err != nil {
		return nil, errors.Wrap(err, "AccessTokens.GetByToken")
	}

	return s.get(ctx, []*sqlf.Query{sqlf.Sprintf("value_sha256=%s", tokenHash)})
}

func (s *accessTokenStore) get(ctx context.Context, conds []*sqlf.Query) (*AccessToken, error) {
	results, err := s.list(ctx, conds, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ErrAccessTokenNotFound
	}
	return results[0], nil
}

// AccessTokensListOptions contains options for listing access tokens.
type AccessTokensListOptions struct {
	SubjectUserID  int32 // only list access tokens with this user as the subject
	LastUsedAfter  *time.Time
	LastUsedBefore *time.Time
	*LimitOffset
}

func (o AccessTokensListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{
		sqlf.Sprintf("deleted_at IS NULL"),
		sqlf.Sprintf("(expires_at IS NULL OR expires_at > NOW())"),
		// We never want internal access tokens to show up in the UI.
		sqlf.Sprintf("internal IS FALSE"),
	}
	if o.SubjectUserID != 0 {
		conds = append(conds, sqlf.Sprintf("subject_user_id=%d", o.SubjectUserID))
	}
	if o.LastUsedAfter != nil {
		conds = append(conds, sqlf.Sprintf("last_used_at>%d", o.LastUsedAfter))
	}
	if o.LastUsedBefore != nil {
		conds = append(conds, sqlf.Sprintf("last_used_at<%d", o.LastUsedBefore))
	}
	return conds
}

func (s *accessTokenStore) List(ctx context.Context, opt AccessTokensListOptions) ([]*AccessToken, error) {
	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s *accessTokenStore) list(ctx context.Context, conds []*sqlf.Query, limitOffset *LimitOffset) ([]*AccessToken, error) {
	q := sqlf.Sprintf(`
SELECT id, subject_user_id, scopes, note, creator_user_id, internal, created_at, last_used_at, expires_at FROM access_tokens
WHERE (%s)
ORDER BY now() - created_at < interval '5 minutes' DESC, -- show recently created tokens first
last_used_at DESC NULLS FIRST, -- ensure newly created tokens show first
created_at DESC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*AccessToken
	for rows.Next() {
		var t AccessToken
		if err := rows.Scan(
			&t.ID,
			&t.SubjectUserID,
			pq.Array(&t.Scopes),
			&t.Note,
			&t.CreatorUserID,
			&t.Internal,
			&t.CreatedAt,
			&t.LastUsedAt,
			&dbutil.NullTime{Time: &t.ExpiresAt},
		); err != nil {
			return nil, err
		}
		results = append(results, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *accessTokenStore) Count(ctx context.Context, opt AccessTokensListOptions) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM access_tokens WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := s.QueryRow(ctx, q).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *accessTokenStore) DeleteByID(ctx context.Context, id int64) error {
	err := s.delete(ctx, sqlf.Sprintf("id=%d", id))
	if err != nil {
		return err
	}

	arg := struct {
		AccessTokenId int64 `json:"access_token_id"`
	}{AccessTokenId: id}

	if err := NewDBWith(s.logger, s).SecurityEventLogs().LogSecurityEvent(ctx, SecurityEventAccessTokenDeleted, "", uint32(actor.FromContext(ctx).UID), "", "BACKEND", arg); err != nil {
		s.logger.Warn("Error logging security event", log.Error(err))
	}

	return nil
}

func (s *accessTokenStore) HardDeleteByID(ctx context.Context, id int64) error {
	res, err := s.ExecResult(ctx, sqlf.Sprintf("DELETE FROM access_tokens WHERE id = %s", id))
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return ErrAccessTokenNotFound
	}

	arg := struct {
		AccessTokenId int64 `json:"access_token_id"`
	}{AccessTokenId: id}

	if err := NewDBWith(s.logger, s).SecurityEventLogs().LogSecurityEvent(ctx, SecurityEventAccessTokenHardDeleted, "", uint32(actor.FromContext(ctx).UID), "", "BACKEND", arg); err != nil {
		s.logger.Warn("Error logging security event", log.Error(err))

	}

	return nil
}

func (s *accessTokenStore) DeleteByToken(ctx context.Context, token string) error {
	tokenHash, err := tokenSHA256Hash(token)
	if err != nil {
		return errors.Wrap(err, "AccessTokens.DeleteByToken")
	}

	err = s.delete(ctx, sqlf.Sprintf("value_sha256=%s", tokenHash))
	if err != nil {
		return err
	}

	arg := struct {
		AccessTokenSHA256 []byte `json:"access_token_sha256"`
	}{
		AccessTokenSHA256: tokenHash,
	}

	if err := NewDBWith(s.logger, s).SecurityEventLogs().LogSecurityEvent(ctx, SecurityEventAccessTokenDeleted, "", uint32(actor.FromContext(ctx).UID), "", "BACKEND", arg); err != nil {
		s.logger.Warn("Error logging security event", log.Error(err))
	}

	return nil
}

func (s *accessTokenStore) delete(ctx context.Context, cond *sqlf.Query) error {
	conds := []*sqlf.Query{cond, sqlf.Sprintf("deleted_at IS NULL")}
	q := sqlf.Sprintf("UPDATE access_tokens SET deleted_at=now() WHERE (%s)", sqlf.Join(conds, ") AND ("))

	res, err := s.ExecResult(ctx, q)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return ErrAccessTokenNotFound
	}
	return nil
}

// tokenSHA256Hash returns the 32-byte long SHA-256 hash of its hex-encoded value
// (after stripping the "sgph_" or "sgp_" token prefix and instance identifier, if present).
func tokenSHA256Hash(token string) ([]byte, error) {
	token, err := accesstoken.ParsePersonalAccessToken(token)
	if err != nil {
		return nil, InvalidTokenError{err}
	}

	value, err := hex.DecodeString(token)
	if err != nil {
		return nil, InvalidTokenError{err}
	}
	return hashutil.ToSHA256Bytes(value), nil
}

type MockAccessTokens struct {
	HardDeleteByID func(id int64) error
}
