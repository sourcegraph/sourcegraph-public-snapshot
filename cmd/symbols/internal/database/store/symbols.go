package store

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func (s *store) CreateSymbolsTable(ctx context.Context) error {
	return s.Exec(ctx, sqlf.Sprintf(`
		CREATE TABLE IF NOT EXISTS symbols (
			name VARCHAR(256) NOT NULL,
			namelowercase VARCHAR(256) NOT NULL,
			path VARCHAR(4096) NOT NULL,
			pathlowercase VARCHAR(4096) NOT NULL,
			line INT NOT NULL,
			kind VARCHAR(255) NOT NULL,
			language VARCHAR(255) NOT NULL,
			parent VARCHAR(255) NOT NULL,
			parentkind VARCHAR(255) NOT NULL,
			signature VARCHAR(255) NOT NULL,
			pattern VARCHAR(255) NOT NULL,
			filelimited BOOLEAN NOT NULL
		)
	`))
}

func (s *store) CreateSymbolIndexes(ctx context.Context) error {
	createIndexQueries := []string{
		`CREATE INDEX idx_name ON symbols(name)`,
		`CREATE INDEX idx_path ON symbols(path)`,
		`CREATE INDEX idx_namelowercase ON symbols(namelowercase)`,
		`CREATE INDEX idx_pathlowercase ON symbols(pathlowercase)`,
	}

	for _, query := range createIndexQueries {
		if err := s.Exec(ctx, sqlf.Sprintf(query)); err != nil {
			return err
		}
	}

	return nil
}

func (s *store) DeletePaths(ctx context.Context, paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	pathQueries := make([]*sqlf.Query, 0, len(paths))
	for _, path := range paths {
		pathQueries = append(pathQueries, sqlf.Sprintf("%s", path))
	}

	return s.Exec(ctx, sqlf.Sprintf(`DELETE FROM symbols WHERE path IN (%s)`, sqlf.Join(pathQueries, ",")))
}

func (s *store) WriteSymbols(ctx context.Context, symbols <-chan result.Symbol) (err error) {
	rows := make(chan []interface{})

	go func() {
		defer close(rows)

		for symbol := range symbols {
			rows <- []interface{}{
				symbol.Name,
				strings.ToLower(symbol.Name),
				symbol.Path,
				strings.ToLower(symbol.Path),
				symbol.Line,
				symbol.Kind,
				symbol.Language,
				symbol.Parent,
				symbol.ParentKind,
				symbol.Signature,
				symbol.Pattern,
				symbol.FileLimited,
			}
		}
	}()

	return batch.InsertValues(
		ctx,
		s.Handle().DB(),
		"symbols",
		batch.MaxNumSQLiteParameters,
		[]string{
			"name",
			"namelowercase",
			"path",
			"pathlowercase",
			"line",
			"kind",
			"language",
			"parent",
			"parentkind",
			"signature",
			"pattern",
			"filelimited",
		},
		rows,
	)
}
