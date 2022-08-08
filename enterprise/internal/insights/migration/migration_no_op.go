package migration

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type migratorNoOp struct {
}

func NewMigratorNoOp() oobmigration.Migrator {
	return &migratorNoOp{}
}

func (m *migratorNoOp) Progress(ctx context.Context) (float64, error) {
	return 1, nil
}

func (m *migratorNoOp) Up(ctx context.Context) (err error) {

	return nil
}

func (m *migratorNoOp) Down(ctx context.Context) (err error) {
	return nil
}
