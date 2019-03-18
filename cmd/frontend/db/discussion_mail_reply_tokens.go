package db

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/dbconn"
)

// discussionMailReplyTokens provides access to the `discussion_mail_reply_tokens` table.
//
// For a detailed overview of the schema, see schema.md.
type discussionMailReplyTokens struct{}

// Generate gets the existing token, or generates a new one, for giving the
// specified user access to the specified thread through only the token.
//
// 🚨 SECURITY: The caller must ensure the token is ONLY given to the user that
// is passed to this method. Anyone with the token has access to reply to the
// specified thread as the specified user, at ANY point in the future.
func (*discussionMailReplyTokens) Generate(ctx context.Context, userID int32, threadID int64) (string, error) {
	if Mocks.DiscussionMailReplyTokens.Generate != nil {
		return Mocks.DiscussionMailReplyTokens.Generate(ctx, userID, threadID)
	}

	// Check if there already exists a token for this userID + threadID pair.
	// If there is, we do not need to store a new one.
	var token string
	err := dbconn.Global.QueryRowContext(ctx, "SELECT token FROM discussion_mail_reply_tokens WHERE user_id=$1 AND thread_id=$2 AND deleted_at IS NULL", userID, threadID).Scan(&token)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if err == nil {
		return token, nil // use the existing token
	}

	// Generate a new secure token and store it. We use SHA256 because it is
	// short and its characters are valid to place in an email address field
	// like "foo+TOKEN@gmail.com", while still providing good security.
	h := sha256.New()
	io.Copy(h, io.LimitReader(cryptorand.Reader, 128)) // Using 128 bytes just to be on the safe side, but 32 bytes should be enough.
	token = fmt.Sprintf("%x", h.Sum(nil))

	_, err = dbconn.Global.ExecContext(ctx, "INSERT INTO discussion_mail_reply_tokens(token, user_id, thread_id) VALUES($1, $2, $3)", token, userID, threadID)
	if err != nil {
		return "", err
	}
	return token, nil
}

// ErrInvalidToken is returned by DiscussionMailReplyTokens.Get when the token is
// invalid.
var ErrInvalidToken = errors.New("invalid token")

// Get returns the user and thread ID found for the given token. If there
// is none, the token is invalid and ErrInvalidToken is returned.
func (*discussionMailReplyTokens) Get(ctx context.Context, token string) (userID int32, threadID int64, err error) {
	if Mocks.DiscussionMailReplyTokens.Get != nil {
		return Mocks.DiscussionMailReplyTokens.Get(ctx, token)
	}
	err = dbconn.Global.QueryRowContext(ctx, "SELECT user_id, thread_id FROM discussion_mail_reply_tokens WHERE token=$1 AND deleted_at IS NULL", token).Scan(
		&userID,
		&threadID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, ErrInvalidToken
		}
		return 0, 0, err
	}
	return userID, threadID, nil
}
