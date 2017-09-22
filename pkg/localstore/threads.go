package localstore

import (
	"context"
	"database/sql"
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

	newThread.CreatedAt = time.Now()
	newThread.UpdatedAt = newThread.CreatedAt
	err := globalDB.QueryRow(
		"INSERT INTO threads(org_repo_id, file, revision, start_line, end_line, start_character, end_character, range_length, created_at, updated_at) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id",
		newThread.OrgRepoID, newThread.File, newThread.Revision, newThread.StartLine, newThread.EndLine, newThread.StartCharacter, newThread.EndCharacter, newThread.RangeLength, newThread.CreatedAt, newThread.UpdatedAt).Scan(&newThread.ID)
	if err != nil {
		return nil, err
	}

	return newThread, nil
}

func (t *threads) Get(ctx context.Context, id int32) (*sourcegraph.Thread, error) {
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

func (t *threads) Update(ctx context.Context, id, repoID int32, archive *bool) (*sourcegraph.Thread, error) {
	if Mocks.Threads.Update != nil {
		return Mocks.Threads.Update(ctx, id, repoID, archive)
	}

	now := time.Now()

	if archive == nil {
		return nil, errors.New("no update values provided")
	}
	if archive != nil {
		archivedAt := pq.NullTime{}
		if *archive == true {
			archivedAt = pq.NullTime{
				Time:  now,
				Valid: true,
			}
		}
		if _, err := globalDB.Exec("UPDATE threads SET archived_at=$1 WHERE id=$2", archivedAt, id); err != nil {
			return nil, err
		}
	}
	if _, err := globalDB.Exec("UPDATE threads SET updated_at=$1 WHERE id=$2", now, id); err != nil {
		return nil, err
	}

	thread, err := t.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return thread, nil
}

func (t *threads) GetByOrg(ctx context.Context, orgID int32, file *string, limit int32) ([]*sourcegraph.Thread, error) {
	if file != nil {
		return t.getBySQL(ctx, "JOIN org_repos ON (org_repos.id = t.org_repo_id) WHERE org_repos.org_id=$1 AND org_repos.deleted_at IS NULL AND t.file=$2 AND t.deleted_at IS NULL LIMIT $3", orgID, file, limit)
	}
	return t.getBySQL(ctx, "JOIN org_repos ON (org_repos.id = t.org_repo_id) WHERE org_repos.org_id=$1 AND org_repos.deleted_at IS NULL AND t.deleted_at IS NULL LIMIT $2", orgID, limit)
}

func (t *threads) GetAllForRepo(ctx context.Context, repoID, limit int32) ([]*sourcegraph.Thread, error) {
	return t.getBySQL(ctx, "WHERE (org_repo_id=$1 AND deleted_at IS NULL) LIMIT $2", repoID, limit)
}

func (t *threads) GetAllForFile(ctx context.Context, repoID int32, file string, limit int32) ([]*sourcegraph.Thread, error) {
	return t.getBySQL(ctx, "WHERE (org_repo_id=$1 AND file=$2 AND deleted_at IS NULL) LIMIT $3", repoID, file, limit)
}

// getBySQL returns threads matching the SQL query, if any exist.
func (*threads) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.Thread, error) {
	rows, err := globalDB.Query("SELECT t.id, t.org_repo_id, t.file, t.revision, t.start_line, t.end_line, t.start_character, t.end_character, t.range_length, t.created_at, t.archived_at FROM threads t "+query, args...)
	if err != nil {
		return nil, err
	}

	threads := []*sourcegraph.Thread{}
	defer rows.Close()
	for rows.Next() {
		var t sourcegraph.Thread
		var archivedAt pq.NullTime
		var rangeLength sql.NullInt64
		err := rows.Scan(&t.ID, &t.OrgRepoID, &t.File, &t.Revision, &t.StartLine, &t.EndLine, &t.StartCharacter, &t.EndCharacter, &rangeLength, &t.CreatedAt, &archivedAt)
		if err != nil {
			return nil, err
		}
		if rangeLength.Valid {
			t.RangeLength = int32(rangeLength.Int64)
		} else {
			t.RangeLength = -1
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
