package store

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func (s *store) CreateSymbolsTable(ctx context.Context) error {
	query := `
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
	`
	return s.Exec(ctx, sqlf.Sprintf(query))
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
	return s.Exec(ctx, sqlf.Sprintf(`DELETE FROM symbols WHERE path = ANY(%s)`, pq.Array(paths)))
}

var symbolsTableColumnNames = []string{
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

	return batch.InsertValues(ctx, s.Handle().DB(), "symbols", symbolsTableColumnNames, rows)
}
