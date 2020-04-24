package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/keegancsmith/sqlf"
)

type PackageModel struct {
	Scheme  string
	Name    string
	Version string
	DumpID  int
}

type ReferenceModel struct {
	Scheme  string
	Name    string
	Version string
	DumpID  int
	Filter  []byte
}

// insertUploads populates the lsif_uploads table with the given upload models.
func insertUploads(t *testing.T, db *sql.DB, uploads ...Upload) {
	for _, upload := range uploads {
		if upload.Commit == "" {
			upload.Commit = makeCommit(upload.ID)
		}
		if upload.State == "" {
			upload.State = "completed"
		}
		if upload.RepositoryID == 0 {
			upload.RepositoryID = 50
		}
		if upload.Indexer == "" {
			upload.Indexer = "lsif-go"
		}

		query := sqlf.Sprintf(`
			INSERT INTO lsif_uploads (
				id,
				commit,
				root,
				visible_at_tip,
				uploaded_at,
				state,
				failure_summary,
				failure_stacktrace,
				started_at,
				finished_at,
				tracing_context,
				repository_id,
				indexer
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
		`,
			upload.ID,
			upload.Commit,
			upload.Root,
			upload.VisibleAtTip,
			upload.UploadedAt,
			upload.State,
			upload.FailureSummary,
			upload.FailureStacktrace,
			upload.StartedAt,
			upload.FinishedAt,
			upload.TracingContext,
			upload.RepositoryID,
			upload.Indexer,
		)

		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while inserting dump: %s", err)
		}
	}
}

// insertPackages populates the lsif_packages table with the given package models.
func insertPackages(t *testing.T, db *sql.DB, packages ...PackageModel) {
	for _, pkg := range packages {
		query := sqlf.Sprintf(`
			INSERT INTO lsif_packages (
				scheme,
				name,
				version,
				dump_id
			) VALUES (%s, %s, %s, %s)
		`,
			pkg.Scheme,
			pkg.Name,
			pkg.Version,
			pkg.DumpID,
		)

		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while inserting package: %s", err)
		}
	}
}

// insertReferences populates the lsif_references table with the given reference models.
func insertReferences(t *testing.T, db *sql.DB, references ...ReferenceModel) {
	for _, reference := range references {
		query := sqlf.Sprintf(`
			INSERT INTO lsif_references (
				scheme,
				name,
				version,
				dump_id,
				filter
			) VALUES (%s, %s, %s, %s, %s)
		`,
			reference.Scheme,
			reference.Name,
			reference.Version,
			reference.DumpID,
			reference.Filter,
		)

		if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error while inserting reference: %s", err)
		}
	}
}

// insertCommits populates the lsif_commits table with the given commit-parent map.
func insertCommits(t *testing.T, db *sql.DB, commits map[string][]string) {
	var values []*sqlf.Query
	for k, vs := range commits {
		if len(vs) == 0 {
			values = append(values, sqlf.Sprintf("(%d, %s, %v)", 50, k, nil))
		}

		for _, v := range vs {
			values = append(values, sqlf.Sprintf("(%d, %s, %s)", 50, k, v))
		}
	}

	query := sqlf.Sprintf(
		"INSERT INTO lsif_commits (repository_id, commit, parent_commit) VALUES %s",
		sqlf.Join(values, ", "),
	)

	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error while inserting commits: %s", err)
	}
}

// getDumpVisibilities returns a map from dump identifiers to its visibility. Fails the test on error.
func getDumpVisibilities(t *testing.T, db *sql.DB) map[int]bool {
	visibilities, err := scanVisibilities(db.Query("SELECT id, visible_at_tip FROM lsif_dumps"))
	if err != nil {
		t.Fatalf("unexpected error while scanning dump visibility: %s", err)
	}

	return visibilities
}

func makeCommit(i int) string {
	return fmt.Sprintf("%040d", i)
}
