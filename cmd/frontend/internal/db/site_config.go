package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

type siteConfig struct{}

var telemetryDisabled = conf.Get().DisableTelemetry

func (o *siteConfig) Get(ctx context.Context) (*sourcegraph.SiteConfig, error) {
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

func (o *siteConfig) getConfiguration(ctx context.Context) (*sourcegraph.SiteConfig, error) {
	configuration := &sourcegraph.SiteConfig{}
	err := globalDB.QueryRowContext(ctx, "SELECT site_id, enable_telemetry, updated_at from site_config LIMIT 1").Scan(
		&configuration.SiteID,
		&configuration.TelemetryEnabled,
		&configuration.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if telemetryDisabled {
		configuration.TelemetryEnabled = false
	}
	return configuration, nil
}

func (o *siteConfig) UpdateConfiguration(ctx context.Context, updatedConfiguration *sourcegraph.SiteConfig) error {
	_, err := o.Get(ctx)
	if err != nil {
		return err
	}
	_, err = globalDB.ExecContext(ctx, "UPDATE site_config SET email=$1, enable_telemetry=$2, updated_at=now()", updatedConfiguration.Email, updatedConfiguration.TelemetryEnabled)
	return err
}

func (o *siteConfig) tryInsertNew(ctx context.Context) error {
	siteID, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	_, err = globalDB.ExecContext(ctx, "INSERT INTO site_config(site_id, enable_telemetry, updated_at) values($1, $2, now())", siteID, !telemetryDisabled)
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
