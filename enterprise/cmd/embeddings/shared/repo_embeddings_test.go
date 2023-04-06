package shared_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

const knnFmtstr = `
	SELECT id
	FROM text_embeddings
	ORDER BY embedding <=> $1::vector
	LIMIT $2
`

const insertEmbeddingFmtstr = `
	INSERT INTO embedding_versions (repo_id, revision)
	VALUES ($1, $2)
`

func TestRestoreEmbeddings(t *testing.T) {
	logger := logtest.NoOp(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	if err := db.Repos().Create(ctx, &types.Repo{Name: "foo"}); err != nil {
		t.Fatal(err)
	}
	repo, err := db.Repos().GetByName(ctx, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Handle().ExecContext(ctx, insertEmbeddingFmtstr, repo.ID, "revisionsha"); err != nil {
		t.Fatal(err)
	}
	var name string
	if err := db.Handle().QueryRowContext(ctx, "SELECT current_database()").Scan(&name); err != nil {
		t.Fatal(err)
	}
	restoreCmd := exec.Command(
		"pg_restore",
		"--format", "custom",
		"--verbose",
		"--data-only",
		"--dbname", name,
		"--table", "code_embeddings",
		"/Users/cbart/embeddings.dump",
	)

	restoreCmd.Stdout = os.Stdout
	restoreCmd.Stderr = os.Stderr

	if err := restoreCmd.Run(); err != nil {
		t.Fatalf("Failed to restore the table dump: %v", err)
	}
	var count int
	if err := db.Handle().QueryRowContext(ctx, "SELECT COUNT(*) FROM code_embeddings").Scan(&count); err != nil {
		t.Fatalf("Failed to query the restored table: %v", err)
	}

	if count == 0 {
		t.Fatal("No records found in the restored table")
	}

	fmt.Printf("Restored %d records in the code_embeddings table\n", count)
	// err := db.WithTransact(ctx, func(tx database.DB) error {
	// 	return nil
	// })
	// if err != nil {
	// 	t.Fatal(err)
	// }
}
