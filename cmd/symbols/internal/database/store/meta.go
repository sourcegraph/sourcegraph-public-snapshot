package store

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

func (s *store) CreateMetaTable(ctx context.Context) error {
	return s.Exec(ctx, sqlf.Sprintf(`
		CREATE TABLE IF NOT EXISTS meta (
			id INTEGER PRIMARY KEY CHECK (id = 0),
			revision TEXT NOT NULL
		)
	`))
}

func (s *store) GetCommit(ctx context.Context) (string, bool, error) {
	return basestore.ScanFirstString(s.Query(ctx, sqlf.Sprintf(`SELECT revision FROM meta`)))
}

func (s *store) InsertMeta(ctx context.Context, commitID string) error {
	return s.Exec(ctx, sqlf.Sprintf(`INSERT INTO meta (id, revision) VALUES (0, %s)`, commitID))
}

func (s *store) UpdateMeta(ctx context.Context, commitID string) error {
	return s.Exec(ctx, sqlf.Sprintf(`UPDATE meta SET revision = %s`, commitID))
}
