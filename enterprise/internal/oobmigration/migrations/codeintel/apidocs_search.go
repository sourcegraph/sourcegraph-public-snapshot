package codeintel

import (
	"context"
	"time"
)

type apiDocsSearchMigrator struct{}

func NewAPIDocsSearchMigrator() TaggedMigrator {
	return &apiDocsSearchMigrator{}
}

func (m *apiDocsSearchMigrator) ID() int                                       { return 12 }
func (m *apiDocsSearchMigrator) Interval() time.Duration                       { return time.Second }
func (m *apiDocsSearchMigrator) Progress(ctx context.Context) (float64, error) { return 1, nil }
func (m *apiDocsSearchMigrator) Up(ctx context.Context) error                  { return nil }
func (m *apiDocsSearchMigrator) Down(ctx context.Context) error                { return nil }
