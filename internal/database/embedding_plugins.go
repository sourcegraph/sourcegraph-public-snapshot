package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type EmbeddingPluginStore interface {
	basestore.ShareableStore
	Get(context.Context, int32) (*types.EmbeddingPlugin, error)
}

type embeddingPluginStore struct {
	logger log.Logger
	*basestore.Store
}

func EmbeddingPluginsWith(logger log.Logger, other basestore.ShareableStore) EmbeddingPluginStore {
	return &embeddingPluginStore{
		logger: logger,
		Store:  basestore.NewWithHandle(other.Handle()),
	}
}

var listEmbeddingPluginsSQL = `
SELECT id, name, original_source_url FROM embedding_plugins
WHERE id = %d;
`

func scanEmbeddingPlugin(logger log.Logger, rows *sql.Rows, p *types.EmbeddingPlugin) error {
	return rows.Scan(
		&p.ID,
		&p.Name,
		&p.OriginalSourceURL,
	)
}

type PluginNotFoundErr struct {
	ID         int32
	Name       string
}

func (e *PluginNotFoundErr) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("plugin not found: name=%q", e.Name)
	}
	if e.ID != 0 {
		return fmt.Sprintf("plugin not found: id=%d", e.ID)
	}
	return "plugin not found"
}

func (s *embeddingPluginStore) Get(ctx context.Context, id int32) (p *types.EmbeddingPlugin, err error) {
	tr, ctx := trace.New(ctx, "embedding_plugins.Get")
	defer tr.FinishWithErr(&err)

	q := sqlf.Sprintf(listEmbeddingPluginsSQL, id)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, &PluginNotFoundErr{ID: id}
	}

	p = &types.EmbeddingPlugin{}
	if err := scanEmbeddingPlugin(s.logger, rows, p); err != nil {
		return nil, err
	}

	return p, nil
}
