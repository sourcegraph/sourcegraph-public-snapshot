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

func (s *sharedItems) Create(ctx context.Context, item *sourcegraph.SharedItem) (*url.URL, error) {
	if item.ULID != "" {
		return nil, errors.New("SharedItems.Create: cannot specify ULID when creating shared item")
	}
	if item.AuthorUserID == "" {
		return nil, errors.New("SharedItems.Create: must specify author user ID")
	}
	if item.ThreadID == nil {
		return nil, errors.New("SharedItems.Create: must specify thread ID")
	}
	if Mocks.SharedItems.Create != nil {
		return Mocks.SharedItems.Create(ctx, item)
	}

	// If a shared item already represents the specified thread, return that
	// shared item instead of creating a new one.
	existingULID, err := s.getByThreadID(ctx, *item.ThreadID, item.Public)
	if err != nil {
		return nil, err
	}
	if existingULID != "" {
		// We already have a shared item for the thread, so do not create another one.
		return s.ulidToURL(existingULID, item.CommentID), nil
	}

	// Generate ULID with entropy from crypto/rand.
	t := time.Now()
	ulid, err := ulid.New(ulid.Timestamp(t), cryptorand.Reader)
	if err != nil {
		return nil, err
	}

	_, err = globalDB.Exec("INSERT INTO shared_items(ulid, author_user_id, thread_id, public) VALUES($1, $2, $3, $4)", ulid.String(), item.AuthorUserID, *item.ThreadID, item.Public)
	if err != nil {
		return nil, err
	}
	return s.ulidToURL(ulid.String(), item.CommentID), nil
}

func (s *sharedItems) Get(ctx context.Context, ulid string) (*sourcegraph.SharedItem, error) {
	if Mocks.SharedItems.Get != nil {
		return Mocks.SharedItems.Get(ctx, ulid)
	}

	item := &sourcegraph.SharedItem{ULID: ulid}
	err := globalDB.QueryRow("SELECT author_user_id, thread_id, comment_id, public FROM shared_items WHERE ulid=$1 AND deleted_at IS NULL", ulid).Scan(
		&item.AuthorUserID,
		&item.ThreadID,
		&item.CommentID,
		&item.Public,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSharedItemNotFound{ulid}
		}
		return nil, err
	}
	return item, nil
}

// getByThreadID gets an existing shared item ULID for the given thread ID.
func (s *sharedItems) getByThreadID(ctx context.Context, threadID int32, wantPublic bool) (string, error) {
	var ulid string
	err := globalDB.QueryRow("SELECT ulid FROM shared_items WHERE thread_id=$1 AND public=$2 AND deleted_at IS NULL", threadID, wantPublic).Scan(
		&ulid,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return ulid, nil
}

// ulidToURL converts the given ulid and optional comment ID until a shared URL.
func (s *sharedItems) ulidToURL(ulid string, commentID *int32) *url.URL {
	shareURL := conf.AppURL.ResolveReference(&url.URL{
		Path: path.Join("c", ulid),
	})
	if commentID != nil {
		// Linking to a comment.
		q := shareURL.Query()
		q.Set("id", fmt.Sprint(*commentID))
		shareURL.RawQuery = q.Encode()
	}
	return shareURL
}
