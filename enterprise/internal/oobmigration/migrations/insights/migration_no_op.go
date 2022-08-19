package insights

import (
	"context"
	"time"
)

type migratorNoOp struct{}

func NewMigratorNoOp() *migratorNoOp {
	return &migratorNoOp{}
}

func (m *migratorNoOp) ID() int                 { return 14 }
func (m *migratorNoOp) Interval() time.Duration { return time.Second * 10 }

func (m *migratorNoOp) Progress(ctx context.Context) (float64, error) { return 1, nil }
func (m *migratorNoOp) Up(ctx context.Context) (err error)            { return nil }
func (m *migratorNoOp) Down(ctx context.Context) (err error)          { return nil }
