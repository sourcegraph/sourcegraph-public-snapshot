package migration

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

// APIDocsSearchMigrationID is a now defunct migration
const APIDocsSearchMigrationID = 12

func NewAPIDocsSearchMigrator(_ int) oobmigration.Migrator {
	return &apiDocsSearchMigrator{}
}

type apiDocsSearchMigrator struct{}

func (m *apiDocsSearchMigrator) Progress(ctx context.Context) (float64, error) { return 1, nil }
func (m *apiDocsSearchMigrator) Up(ctx context.Context) error                  { return nil }
func (m *apiDocsSearchMigrator) Down(ctx context.Context) error                { return nil }
