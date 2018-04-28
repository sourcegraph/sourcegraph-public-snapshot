package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
)

type siteConfig struct{}

func (o *siteConfig) Get(ctx context.Context) (*types.SiteConfig, error) {
	if Mocks.SiteConfig.Get != nil {
		return Mocks.SiteConfig.Get(ctx)
	}

	configuration, err := o.getConfiguration(ctx)
	if err == nil {
		return configuration, nil
	}
	err = o.tryInsertNew(ctx)
	if err != nil {
		return nil, err
	}
	return o.getConfiguration(ctx)
}

func (o *siteConfig) getConfiguration(ctx context.Context) (*types.SiteConfig, error) {
	configuration := &types.SiteConfig{}
	err := globalDB.QueryRowContext(ctx, "SELECT site_id, initialized FROM site_config LIMIT 1").Scan(
		&configuration.SiteID,
		&configuration.Initialized,
	)
	return configuration, err
}

func (o *siteConfig) tryInsertNew(ctx context.Context) error {
	siteID, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	_, err = globalDB.ExecContext(ctx, "INSERT INTO site_config(site_id) values($1)", siteID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Constraint == "site_config_pkey" {
				// The row we were trying to insert already exists.
				// Don't treat this as an error.
				err = nil
			}

		}
	}
	return err
}
