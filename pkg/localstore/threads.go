package localstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"

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
	var err error
	if newThread.Lines == nil {
		err = globalDB.QueryRow(
			"INSERT INTO threads(org_repo_id, file, revision, branch, start_line, end_line, start_character, end_character, range_length, created_at, updated_at, author_user_id) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id",
			newThread.OrgRepoID, newThread.File, newThread.Revision, newThread.Branch, newThread.StartLine, newThread.EndLine, newThread.StartCharacter, newThread.EndCharacter, newThread.RangeLength, newThread.CreatedAt, newThread.UpdatedAt, newThread.AuthorUserID).Scan(&newThread.ID)
	} else {
		err = globalDB.QueryRow(
			"INSERT INTO threads(org_repo_id, file, revision, branch, start_line, end_line, start_character, end_character, range_length, created_at, updated_at, author_user_id, html_lines_before, html_lines, html_lines_after, text_lines_before, text_lines, text_lines_after, text_lines_selection_range_start, text_lines_selection_range_length) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $18, $19, $20) RETURNING id",
			newThread.OrgRepoID, newThread.File, newThread.Revision, newThread.Branch, newThread.StartLine, newThread.EndLine, newThread.StartCharacter, newThread.EndCharacter, newThread.RangeLength, newThread.CreatedAt, newThread.UpdatedAt, newThread.AuthorUserID, newThread.Lines.HTMLBefore, newThread.Lines.HTML, newThread.Lines.HTMLAfter, newThread.Lines.TextBefore, newThread.Lines.Text, newThread.Lines.TextAfter, newThread.Lines.TextSelectionRangeStart, newThread.Lines.TextSelectionRangeLength).Scan(&newThread.ID)
	}
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

func (t *threads) listQuery(ctx context.Context, repoID, orgID *int32, branch, file *string) *sqlf.Query {
	var join string
	conds := []*sqlf.Query{}
	if repoID != nil {
		conds = append(conds, sqlf.Sprintf("t.org_repo_id=%d", repoID))
	}
	if orgID != nil {
		join = "JOIN org_repos ON (org_repos.id = t.org_repo_id) "
		conds = append(conds, sqlf.Sprintf("(org_repos.org_id=%d AND org_repos.deleted_at IS NULL)", *orgID))
	}
	if branch != nil {
		conds = append(conds, sqlf.Sprintf("t.branch=%s", *branch))
	}
	if file != nil {
		conds = append(conds, sqlf.Sprintf("t.file=%s", *file))
	}
	conds = append(conds, sqlf.Sprintf("t.deleted_at IS NULL"))
	return sqlf.Sprintf(join+"WHERE %s", sqlf.Join(conds, "AND"))
}

func (t *threads) List(ctx context.Context, repoID, orgID *int32, branch, file *string, limit int32) ([]*sourcegraph.Thread, error) {
	q := sqlf.Sprintf("%s LIMIT %d", t.listQuery(ctx, repoID, orgID, branch, file), limit)
	return t.getBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
}

func (t *threads) Count(ctx context.Context, repoID, orgID *int32, branch, file *string, limit int32) (int32, error) {
	q := t.listQuery(ctx, repoID, orgID, branch, file)
	return t.getCountBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
}

func (t *threads) getCountBySQL(ctx context.Context, query string, args ...interface{}) (int32, error) {
	var count int32
	rows := globalDB.QueryRow("SELECT count(*) FROM threads t "+query, args...)
	err := rows.Scan(&count)
	return count, err
}

// getBySQL returns threads matching the SQL query, if any exist.
func (*threads) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.Thread, error) {
	rows, err := globalDB.Query("SELECT t.id, t.org_repo_id, t.file, t.revision, t.branch, t.start_line, t.end_line, t.start_character, t.end_character, t.range_length, t.created_at, t.updated_at, t.archived_at, t.author_user_id, t.html_lines_before, t.html_lines, t.html_lines_after, t.text_lines_before, t.text_lines, t.text_lines_after, t.text_lines_selection_range_start, t.text_lines_selection_range_length FROM threads t "+query, args...)
	if err != nil {
		return nil, err
	}

	threads := []*sourcegraph.Thread{}
	defer rows.Close()
	for rows.Next() {
		var t sourcegraph.Thread
		var archivedAt pq.NullTime
		var rangeLength sql.NullInt64
		var authorUserID, htmlBefore, html, htmlAfter, textBefore, text, textAfter sql.NullString
		var textSelectionRangeStart, textSelectionRangeLength sql.NullInt64
		err := rows.Scan(&t.ID, &t.OrgRepoID, &t.File, &t.Revision, &t.Branch, &t.StartLine, &t.EndLine, &t.StartCharacter, &t.EndCharacter, &rangeLength, &t.CreatedAt, &t.UpdatedAt, &archivedAt, &authorUserID, &htmlBefore, &html, &htmlAfter, &textBefore, &text, &textAfter, &textSelectionRangeStart, &textSelectionRangeLength)
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
		if authorUserID.Valid {
			t.AuthorUserID = authorUserID.String
		}
		if htmlBefore.Valid && html.Valid && htmlAfter.Valid && textBefore.Valid && text.Valid && textAfter.Valid && textSelectionRangeStart.Valid && textSelectionRangeLength.Valid {
			t.Lines = &sourcegraph.ThreadLines{
				HTMLBefore:               htmlBefore.String,
				HTML:                     html.String,
				HTMLAfter:                htmlAfter.String,
				TextBefore:               textBefore.String,
				Text:                     text.String,
				TextAfter:                textAfter.String,
				TextSelectionRangeLength: int32(textSelectionRangeLength.Int64),
				TextSelectionRangeStart:  int32(textSelectionRangeStart.Int64),
			}
		}

		threads = append(threads, &t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return threads, nil
}
