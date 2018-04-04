package db

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
)

// AccessToken describes an access token. The actual token (that a caller must supply to
// authenticate) is not stored and is not present in this struct.
type AccessToken struct {
	ID         int64
	UserID     int32
	Note       string
	CreatedAt  time.Time
	LastUsedAt *time.Time
}

// ErrAccessTokenNotFound occurs when a database operation expects a specific access token to exist
// but it does not exist.
var ErrAccessTokenNotFound = errors.New("access token not found")

// accessTokens implements autocert.Cache
type accessTokens struct{}

// Create creates an access token for the specified user. The secret token value itself is
// returned. The caller is responsible for presenting this value to the end user; Sourcegraph does
// not retain it (only a hash of it).
//
// The secret token value is a long random string; it is what API clients must provide to
// authenticate their requests. We store the SHA-256 hash of the secret token value in the
// database. This lets us verify a token's validity (in the (*accessTokens).Lookup method) quickly,
// while still ensuring that an attacker who obtains the access_tokens DB table would not be able to
// impersonate a token holder. We don't use bcrypt because the original secret is a randomly
// generated string (not a password), so it's implausible for an attacker to brute-force the input
// space; also bcrypt is slow and would add noticeable latency to each request that supplied a
// token.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to create tokens for the
// specified user (i.e., that the actor is either the user or a site admin).
func (s *accessTokens) Create(ctx context.Context, userID int32, note string) (id int64, token string, err error) {
	if Mocks.AccessTokens.Create != nil {
		return Mocks.AccessTokens.Create(userID, note)
	}

	var b [20]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, "", err
	}
	token = hex.EncodeToString(b[:])

	if err := globalDB.QueryRowContext(ctx,
		"INSERT INTO access_tokens(user_id, value_sha256, note) VALUES($1, $2, $3) RETURNING id",
		userID, toSHA256Bytes(b[:]), note,
	).Scan(&id); err != nil {
		return 0, "", err
	}
	return id, token, nil
}

// Lookup looks up the access token. If it's valid, it returns the owner's user ID. Otherwise
// ErrAccessTokenNotFound is returned.
//
// Calling Lookup also updates the access token's last-used-at date.
//
// 🚨 SECURITY: This returns a user ID if and only if the tokenHexEncoded corresponds to a valid,
// non-deleted access token.
func (s *accessTokens) Lookup(ctx context.Context, tokenHexEncoded string) (userID int32, err error) {
	if Mocks.AccessTokens.Lookup != nil {
		return Mocks.AccessTokens.Lookup(tokenHexEncoded)
	}

	token, err := hex.DecodeString(tokenHexEncoded)
	if err != nil {
		return 0, errors.Wrap(err, "AccessTokens.Lookup")
	}

	if err := globalDB.QueryRowContext(ctx,
		"UPDATE access_tokens SET last_used_at=now() WHERE value_sha256=$1 AND deleted_at IS NULL RETURNING user_id",
		toSHA256Bytes(token),
	).Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrAccessTokenNotFound
		}
		return 0, err
	}
	return userID, nil
}

// GetByID retrieves the access token (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this access token.
func (s *accessTokens) GetByID(ctx context.Context, id int64) (*AccessToken, error) {
	if Mocks.AccessTokens.GetByID != nil {
		return Mocks.AccessTokens.GetByID(id)
	}

	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)}, nil)
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
	UserID int32 // only list access tokens for the user
	*LimitOffset
}

func (o AccessTokensListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("deleted_at IS NULL")}
	if o.UserID != 0 {
		conds = append(conds, sqlf.Sprintf("user_id=%d", o.UserID))
	}
	return conds
}

// List lists all access tokens that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s *accessTokens) List(ctx context.Context, opt AccessTokensListOptions) ([]*AccessToken, error) {
	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s *accessTokens) list(ctx context.Context, conds []*sqlf.Query, limitOffset *LimitOffset) ([]*AccessToken, error) {
	q := sqlf.Sprintf(`
SELECT id, user_id, note, created_at, last_used_at FROM access_tokens
WHERE (%s)
ORDER BY now() - created_at < interval '5 minutes' DESC, -- show recently created tokens first
last_used_at DESC NULLS FIRST, -- ensure newly created tokens show first
created_at DESC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)

	rows, err := globalDB.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*AccessToken
	for rows.Next() {
		var t AccessToken
		if err := rows.Scan(&t.ID, &t.UserID, &t.Note, &t.CreatedAt, &t.LastUsedAt); err != nil {
			return nil, err
		}
		results = append(results, &t)
	}
	return results, nil
}

// Count counts all access tokens that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the tokens.
func (s *accessTokens) Count(ctx context.Context, opt AccessTokensListOptions) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM access_tokens WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := globalDB.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// DeleteByID deletes an access token given its ID and associated user.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to delete the token.
func (s *accessTokens) DeleteByID(ctx context.Context, id int64, userID int32) error {
	if Mocks.AccessTokens.DeleteByID != nil {
		return Mocks.AccessTokens.DeleteByID(id, userID)
	}
	return s.delete(ctx, sqlf.Sprintf("id=%d AND user_id=%d", id, userID))
}

// DeleteByToken deletes an access token given the secret token value itself (i.e., the same value
// that an API client would use to authenticate).
func (s *accessTokens) DeleteByToken(ctx context.Context, tokenHexEncoded string) error {
	token, err := hex.DecodeString(tokenHexEncoded)
	if err != nil {
		return errors.Wrap(err, "AccessTokens.DeleteByToken")
	}
	return s.delete(ctx, sqlf.Sprintf("value_sha256=%s", toSHA256Bytes(token)))
}

func (s *accessTokens) delete(ctx context.Context, cond *sqlf.Query) error {
	conds := []*sqlf.Query{cond, sqlf.Sprintf("deleted_at IS NULL")}
	q := sqlf.Sprintf("UPDATE access_tokens SET deleted_at=now() WHERE (%s)", sqlf.Join(conds, ") AND ("))

	res, err := globalDB.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
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

func toSHA256Bytes(input []byte) []byte {
	b := sha256.Sum256(input)
	return b[:]
}

type MockAccessTokens struct {
	Create     func(userID int32, note string) (id int64, token string, err error)
	DeleteByID func(id int64, userID int32) error
	Lookup     func(tokenHexEncoded string) (userID int32, err error)
	GetByID    func(id int64) (*AccessToken, error)
}
