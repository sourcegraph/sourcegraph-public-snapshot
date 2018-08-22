package db

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// TODO(slimsag:discussions): Remove this entire file once dogfood is migrated.

func (t *discussionThreads) dogfoodMigration() {
	if !strings.Contains(globals.AppURL.String(), "sourcegraph.sgdev.org") {
		return
	}

	ctx := context.Background()

	// Begin transaction and lock all tables to avoid any races or bad migrations.
	tx, err := globalDB.Begin()
	if err != nil {
		log15.Error("dogfood migration (globalDB.Begin)", "error", err)
		return
	}
	_, err = tx.ExecContext(ctx, "LOCK TABLE discussion_comments IN EXCLUSIVE MODE;")
	if err != nil {
		log15.Error("dogfood migration (lock 1)", "error", err)
		tx.Rollback()
		return
	}
	_, err = tx.ExecContext(ctx, "LOCK TABLE discussion_threads IN EXCLUSIVE MODE;")
	if err != nil {
		log15.Error("dogfood migration (lock 1)", "error", err)
		tx.Rollback()
		return
	}
	_, err = tx.ExecContext(ctx, "LOCK TABLE discussion_threads_target_repo IN EXCLUSIVE MODE;")
	if err != nil {
		log15.Error("dogfood migration (lock 1)", "error", err)
		tx.Rollback()
		return
	}

	q := sqlf.Sprintf("ORDER BY id DESC")
	threads, err := t.getBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		log15.Error("dogfood migration (getBySQL)", "error", err)
		tx.Rollback()
		return
	}

	var threadsNeedingMigration []*types.DiscussionThread
	for _, thread := range threads {
		tr := thread.TargetRepo
		if (*tr.LinesBefore == "" && *tr.Lines == "" && *tr.LinesAfter == "") || *tr.EndLine == 0 {
			threadsNeedingMigration = append(threadsNeedingMigration, thread)
		}
	}
	threads = threadsNeedingMigration
	if len(threads) == 0 {
		tx.Rollback()
		return
	}
	log15.Warn("dogfood migration of discussions found threads needing migration", "found", len(threads))

	// Migrate each thread.
	ok := 0
	for _, thread := range threads {
		log15.Warn("dogfood migration: migrating discussion thread", "id", thread.ID, "title", thread.Title)
		if thread.TargetRepo == nil {
			log15.Error("dogfood migration", "error", "target repo missing")
			continue
		}
		if thread.TargetRepo.Revision == nil {
			log15.Error("dogfood migration", "error", "target repo revision missing")
			continue
		}
		if thread.TargetRepo.Path == nil {
			log15.Error("dogfood migration", "error", "target repo path missing")
			continue
		}
		if !thread.TargetRepo.HasSelection() {
			log15.Error("dogfood migration", "error", "target repo selection missing")
			continue
		}

		// Find the lines we will insert.
		repo, err := Repos.Get(ctx, thread.TargetRepo.RepoID)
		if err != nil {
			log15.Error("dogfood migration (Repos.Get)", "error", err)
			continue
		}
		contents, err := git.ReadFile(ctx, gitserver.Repo{Name: repo.URI}, api.CommitID(*thread.TargetRepo.Revision), *thread.TargetRepo.Path)
		if err != nil {
			log15.Error("dogfood migration (git.ReadFIle)", "error", err)
			continue
		}

		// Previously, endLine was incorrectly 0 when not specified.
		endLine := *thread.TargetRepo.EndLine
		if endLine == 0 {
			endLine = *thread.TargetRepo.StartLine + 1
		}

		linesBefore, lines, linesAfter := LinesForSelection(string(contents), LineRange{
			StartLine: int(*thread.TargetRepo.StartLine),
			EndLine:   int(endLine),
		})

		if _, err := tx.ExecContext(ctx, "UPDATE discussion_threads_target_repo SET end_line=$1 WHERE id=$2", endLine, thread.ID); err != nil {
			log15.Error("dogfood migration (update end_line)", "error", err)
			continue
		}
		if _, err := tx.ExecContext(ctx, "UPDATE discussion_threads_target_repo SET lines_before=$1 WHERE id=$2", linesBefore, thread.ID); err != nil {
			log15.Error("dogfood migration (update lines_before)", "error", err)
			continue
		}
		if _, err := tx.ExecContext(ctx, "UPDATE discussion_threads_target_repo SET lines=$1 WHERE id=$2", lines, thread.ID); err != nil {
			log15.Error("dogfood migration (update lines)", "error", err)
			continue
		}
		if _, err := tx.ExecContext(ctx, "UPDATE discussion_threads_target_repo SET lines_after=$1 WHERE id=$2", linesAfter, thread.ID); err != nil {
			log15.Error("dogfood migration (update lines_after)", "error", err)
			continue
		}
		ok++
	}

	if err := tx.Commit(); err != nil {
		log15.Error("dogfood migration (commit)", "error", err)
	}
	log15.Warn("dogfood migration of discussions complete", "migrated", ok, "of", len(threads))
	/*
		select count(*) from discussion_threads_target_repo WHERE lines_before = '' AND lines = '' AND lines_after = '';

		select thread_id, start_line, end_line, lines_before, lines, lines_after from discussion_threads_target_repo WHERE lines_before = '' AND lines = '' AND lines_after = '';
	*/
}

// LineRange represents a line range in a file.
type LineRange struct {
	// StarLine of the range (zero-based, inclusive).
	StartLine int

	// EndLine of the range (zero-based, exclusive).
	EndLine int
}

// LinesForSelection returns the lines from the given file's contents for the
// given selection.
func LinesForSelection(fileContent string, selection LineRange) (linesBefore, lines, linesAfter string) {
	allLines := strings.Split(fileContent, "\n")
	clamp := func(v, min, max int) int {
		if v < min {
			return min
		} else if v > max {
			return max
		}
		return v
	}
	linesForRange := func(startLine, endLine int) string {
		startLine = clamp(startLine, 0, len(allLines))
		endLine = clamp(endLine, 0, len(allLines))
		selectedLines := allLines[startLine:endLine]
		l := strings.Join(selectedLines, "\n")
		if len(selectedLines) != 0 && endLine != len(allLines) {
			l += "\n"
		}
		return l
	}
	linesBefore = linesForRange(selection.StartLine-3, selection.StartLine)
	lines = linesForRange(selection.StartLine, selection.EndLine)
	linesAfter = linesForRange(selection.EndLine, selection.EndLine+3)
	return
}
