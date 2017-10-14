package localstore

import (
	"context"
	cryptorand "crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path"
	"time"

	"github.com/oklog/ulid"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

// ErrSharedItemNotFound is an error returned by SharedItems.Get when the
// requested shared item is not found.
type ErrSharedItemNotFound struct {
	ulid string
}

func (err ErrSharedItemNotFound) Error() string {
	return fmt.Sprintf("shared item not found: %q", err.ulid)
}

// sharedItems provides access to the `shared_items` table.
//
// For a detailed overview of the schema, see schema.md.
type sharedItems struct{}

func (s *sharedItems) Create(ctx context.Context, item *sourcegraph.SharedItem) (string, error) {
	if item.ULID != "" {
		return "", errors.New("SharedItems.Create: cannot specify ULID when creating shared item")
	}
	if item.AuthorUserID == "" {
		return "", errors.New("SharedItems.Create: must specify author user ID")
	}
	if Mocks.SharedItems.Create != nil {
		return Mocks.SharedItems.Create(ctx, item)
	}

	// Generate ULID with entropy from crypto/rand.
	t := time.Now()
	ulid, err := ulid.New(ulid.Timestamp(t), cryptorand.Reader)
	if err != nil {
		return "", err
	}

	switch {
	case item.ThreadID != nil && item.CommentID == nil:
		_, err = globalDB.Exec("INSERT INTO shared_items(ulid, author_user_id, thread_id) VALUES($1, $2, $3)", ulid.String(), item.AuthorUserID, *item.ThreadID)
	case item.ThreadID == nil && item.CommentID != nil:
		_, err = globalDB.Exec("INSERT INTO shared_items(ulid, author_user_id, comment_id) VALUES($1, $2, $3)", ulid.String(), item.AuthorUserID, *item.CommentID)
	default:
		return "", errors.New("SharedItems.Create: invalid shared item (expected exactly one of ThreadID or CommentID)")
	}
	if err != nil {
		return "", err
	}
	shareURL := conf.AppURL.ResolveReference(&url.URL{Path: path.Join("c", ulid.String())})
	return shareURL.String(), nil
}

func (s *sharedItems) Get(ctx context.Context, ulid string) (*sourcegraph.SharedItem, error) {
	if Mocks.SharedItems.Get != nil {
		return Mocks.SharedItems.Get(ctx, ulid)
	}

	item := &sourcegraph.SharedItem{ULID: ulid}
	err := globalDB.QueryRow("SELECT author_user_id, thread_id, comment_id FROM shared_items WHERE ulid=$1 AND deleted_at IS NULL", ulid).Scan(
		&item.AuthorUserID,
		&item.ThreadID,
		&item.CommentID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSharedItemNotFound{ulid}
		}
		return nil, err
	}
	return item, nil
}
