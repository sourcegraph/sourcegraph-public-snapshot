package migration

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type apiDocsSearchMigrator struct{}

func NewAPIDocsSearchMigrator(_ int) oobmigration.Migrator {
	return &apiDocsSearchMigrator{}
}

func (m *apiDocsSearchMigrator) ID() int                                       { return 12 }
func (m *apiDocsSearchMigrator) Interval() time.Duration                       { return time.Second }
func (m *apiDocsSearchMigrator) Progress(ctx context.Context) (float64, error) { return 1, nil }
func (m *apiDocsSearchMigrator) Up(ctx context.Context) error                  { return nil }
func (m *apiDocsSearchMigrator) Down(ctx context.Context) error                { return nil }
