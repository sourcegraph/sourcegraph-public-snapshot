package db

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

// AddTarget adds a target to a thread. A thread has zero or more targets.
func (t *discussionThreads) AddTarget(ctx context.Context, tr *types.DiscussionThreadTargetRepo) (*types.DiscussionThreadTargetRepo, error) {
	if rev := tr.Revision; rev != nil {
		if !git.IsAbsoluteRevision(*rev) {
			return nil, errors.New("thread target revision must be an absolute Git revision (40 character SHA-1 hash)")
		}
	}

	var fields []*sqlf.Query
	var values []*sqlf.Query
	field := func(name string, arg interface{}) {
		fields = append(fields, sqlf.Sprintf("%s", sqlf.Sprintf(name)))
		values = append(values, sqlf.Sprintf("%v", arg))
	}
	field("thread_id", tr.ThreadID)
	field("repo_id", tr.RepoID)
	if tr.Path != nil {
		field("path", *tr.Path)
	}
	if tr.Branch != nil {
		field("branch", *tr.Branch)
	}
	if tr.Revision != nil {
		field("revision", *tr.Revision)
	}
	if tr.HasSelection() {
		field("start_line", *tr.StartLine)
		field("end_line", *tr.EndLine)
		field("start_character", *tr.StartCharacter)
		field("end_character", *tr.EndCharacter)
		field("lines_before", strings.Join(*tr.LinesBefore, "\n"))
		field("lines", strings.Join(*tr.Lines, "\n"))
		field("lines_after", strings.Join(*tr.LinesAfter, "\n"))
	}
	q := sqlf.Sprintf("INSERT INTO discussion_threads_target_repo(%v) VALUES (%v) RETURNING id", sqlf.Join(fields, ",\n"), sqlf.Join(values, ","))

	// To debug query building, uncomment these lines:
	// fmt.Println(q.Query(sqlf.PostgresBindVar))
	// fmt.Println(q.Args())

	err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&tr.ID)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

// RemoveTarget removes a target from a thread.
func (t *discussionThreads) RemoveTarget(ctx context.Context, targetID int64) error {
	_, err := dbconn.Global.ExecContext(ctx, `DELETE FROM discussion_threads_target_repo WHERE id=$1`, targetID)
	return err
}

// ListTargets returns a list of targets for a thread.
func (t *discussionThreads) ListTargets(ctx context.Context, threadID int64) ([]*types.DiscussionThreadTargetRepo, error) {
	if Mocks.DiscussionThreads.ListTargets != nil {
		return Mocks.DiscussionThreads.ListTargets(threadID)
	}

	rows, err := dbconn.Global.QueryContext(ctx, `
		SELECT
			t.id,
			t.thread_id,
			t.repo_id,
			t.path,
			t.branch,
			t.revision,
			t.start_line,
			t.end_line,
			t.start_character,
			t.end_character,
			t.lines_before,
			t.lines,
			t.lines_after
		FROM discussion_threads_target_repo t WHERE thread_id=$1
	`, threadID)
	if err != nil {
		return nil, err
	}

	targets := []*types.DiscussionThreadTargetRepo{}
	defer rows.Close()
	for rows.Next() {
		var (
			tr                             types.DiscussionThreadTargetRepo
			linesBefore, lines, linesAfter *string
		)
		err := rows.Scan(
			&tr.ID,
			&tr.ThreadID,
			&tr.RepoID,
			&tr.Path,
			&tr.Branch,
			&tr.Revision,
			&tr.StartLine,
			&tr.EndLine,
			&tr.StartCharacter,
			&tr.EndCharacter,
			&linesBefore,
			&lines,
			&linesAfter,
		)
		if err != nil {
			return nil, err
		}
		if linesBefore != nil {
			linesBeforeSplit := strings.Split(*linesBefore, "\n")
			tr.LinesBefore = &linesBeforeSplit
		}
		if lines != nil {
			linesSplit := strings.Split(*lines, "\n")
			tr.Lines = &linesSplit
		}
		if linesAfter != nil {
			linesAfterSplit := strings.Split(*linesAfter, "\n")
			tr.LinesAfter = &linesAfterSplit
		}
		targets = append(targets, &tr)
	}
	return targets, rows.Err()
}
