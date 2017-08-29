package localstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// threads provides access to the `threads` table.
//
// For a detailed overview of the schema, see schema.txt.
type threads struct{}

func (*threads) Create(ctx context.Context, newThread *sourcegraph.Thread) (*sourcegraph.Thread, error) {
	if Mocks.Threads.Create != nil {
		return Mocks.Threads.Create(ctx, newThread)
	}

	if newThread == nil {
		return nil, errors.New("error creating thread: newThread is nil")
	}

	newThread.UpdatedAt = time.Now()
	newThread.UpdatedAt = newThread.CreatedAt
	err := appDBH(ctx).QueryRow(
		"INSERT INTO threads(local_repo_id, file, revision, start_line, end_line, start_character, end_character, created_at, updated_at) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
		newThread.LocalRepoID, newThread.File, newThread.Revision, newThread.StartLine, newThread.EndLine, newThread.StartCharacter, newThread.EndCharacter, newThread.CreatedAt, newThread.UpdatedAt).Scan(&newThread.ID)
	if err != nil {
		return nil, err
	}

	return newThread, nil
}

func (t *threads) Get(ctx context.Context, id int64) (*sourcegraph.Thread, error) {
	if Mocks.Threads.Get != nil {
		return Mocks.Threads.Get(ctx, id)
	}

	threads, err := t.getBySQL(ctx, "WHERE (id=$1 AND deleted_at IS NULL) LIMIT 1", id)
	if err != nil {
		return nil, err
	}
	if len(threads) == 0 {
		return nil, fmt.Errorf("thread %d not found", id)
	}
	return threads[0], nil
}

func (t *threads) Update(ctx context.Context, id, repoID int64, archive *bool) (*sourcegraph.Thread, error) {
	if Mocks.Threads.Update != nil {
		return Mocks.Threads.Update(ctx, id, repoID, archive)
	}

	if archive == nil {
		return nil, errors.New("no update values provided")
	}
	if archive != nil {
		archivedAt := pq.NullTime{}
		if *archive == true {
			archivedAt = pq.NullTime{
				Time:  time.Now(),
				Valid: true,
			}
		}
		if _, err := appDBH(ctx).Exec("UPDATE threads SET archived_at=$1 WHERE id=$2", archivedAt, id); err != nil {
			return nil, err
		}
	}
	if _, err := appDBH(ctx).Exec("UPDATE threads SET updated_at=$1 WHERE id=$2", time.Now(), id); err != nil {
		return nil, err
	}

	thread, err := t.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return thread, nil
}

func (t *threads) GetAllForFile(ctx context.Context, repoID int64, file string) ([]*sourcegraph.Thread, error) {
	return t.getBySQL(ctx, "WHERE (local_repo_id=$1 AND file=$2 AND deleted_at IS NULL)", repoID, file)
}

// getBySQL returns threads matching the SQL query, if any exist.
func (*threads) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.Thread, error) {
	rows, err := appDBH(ctx).Query("SELECT id, local_repo_id, file, revision, start_line, end_line, start_character, end_character, created_at, archived_at FROM threads "+query, args...)
	if err != nil {
		return nil, err
	}

	threads := []*sourcegraph.Thread{}
	defer rows.Close()
	for rows.Next() {
		var t sourcegraph.Thread
		var archivedAt pq.NullTime
		err := rows.Scan(&t.ID, &t.LocalRepoID, &t.File, &t.Revision, &t.StartLine, &t.EndLine, &t.StartCharacter, &t.EndCharacter, &t.CreatedAt, &archivedAt)
		if err != nil {
			return nil, err
		}

		if archivedAt.Valid {
			t.ArchivedAt = &archivedAt.Time
		} else {
			t.ArchivedAt = nil
		}

		threads = append(threads, &t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return threads, nil
}
