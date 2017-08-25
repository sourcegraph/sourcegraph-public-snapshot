package localstore

import (
	"context"
	"errors"
	"fmt"
	"time"

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

	newThread.CreatedAt = time.Now()
	err := appDBH(ctx).QueryRow(
		"INSERT INTO threads(local_repo_id, file, revision, start_line, end_line, start_character, end_character, created_at) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		newThread.LocalRepoID, newThread.File, newThread.Revision, newThread.StartLine, newThread.EndLine, newThread.StartCharacter, newThread.EndCharacter, newThread.CreatedAt).Scan(&newThread.ID)
	if err != nil {
		return nil, err
	}

	return newThread, nil
}

func (t *threads) Get(ctx context.Context, id int64) (*sourcegraph.Thread, error) {
	if Mocks.Threads.Get != nil {
		return Mocks.Threads.Get(ctx, id)
	}

	threads, err := t.getBySQL(ctx, "WHERE (id=$1) LIMIT 1", id)
	if err != nil {
		return nil, err
	}
	if len(threads) == 0 {
		return nil, fmt.Errorf("thread %d not found", id)
	}
	return threads[0], nil
}

func (t *threads) GetAllForFile(ctx context.Context, repoID int64, file string) ([]*sourcegraph.Thread, error) {
	return t.getBySQL(ctx, "WHERE (local_repo_id=$1 AND file=$2)", repoID, file)
}

// getBySQL returns threads matching the SQL query, if any exist.
func (*threads) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.Thread, error) {
	rows, err := appDBH(ctx).Query("SELECT id, local_repo_id, file, revision, start_line, end_line, start_character, end_character, created_at FROM threads "+query, args...)
	if err != nil {
		return nil, err
	}

	threads := []*sourcegraph.Thread{}
	defer rows.Close()
	for rows.Next() {
		var t sourcegraph.Thread
		err := rows.Scan(&t.ID, &t.LocalRepoID, &t.File, &t.Revision, &t.StartLine, &t.EndLine, &t.StartCharacter, &t.EndCharacter, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		threads = append(threads, &t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return threads, nil
}
