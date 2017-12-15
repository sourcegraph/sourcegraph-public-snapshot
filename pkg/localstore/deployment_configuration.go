package localstore

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

type deploymentConfiguration struct{}

var isDisabled = env.Get("DISABLE_TELEMETRY", "false", "disable telemetry")

func (o *deploymentConfiguration) Get(ctx context.Context) (*sourcegraph.DeploymentConfiguration, error) {
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

func (o *deploymentConfiguration) getConfiguration(ctx context.Context) (*sourcegraph.DeploymentConfiguration, error) {
	configuration := &sourcegraph.DeploymentConfiguration{}
	err := globalDB.QueryRowContext(ctx, "SELECT app_id, enable_telemetry, last_updated from deployment_configuration LIMIT 1").Scan(
		&configuration.AppID,
		&configuration.TelemetryEnabled,
		&configuration.LastUpdated,
	)
	if err != nil {
		return nil, err
	}
	return configuration, nil
}

func (o *deploymentConfiguration) UpdateConfiguration(ctx context.Context, updatedConfiguration *sourcegraph.DeploymentConfiguration) error {
	_, err := o.Get(ctx)
	if err != nil {
		return err
	}
	t := time.Now()
	_, err = globalDB.ExecContext(ctx, "UPDATE deployment_configuration SET email = $1, enable_telemetry = $2, last_updated = $3 where id = 1", updatedConfiguration.Email, updatedConfiguration.TelemetryEnabled, t.String())
	return err
}

func (o *deploymentConfiguration) tryInsertNew(ctx context.Context) error {
	appID, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	var lastUpdated = ""
	telemetryDisabled, err := strconv.ParseBool(isDisabled)
	if err != nil {
		return err
	}
	if telemetryDisabled {
		lastUpdated = time.Now().String()
	}
	_, err = globalDB.ExecContext(ctx, "INSERT INTO deployment_configuration(id, app_id, enable_telemetry, last_updated) values(1, $1, $2, $3)", appID, !telemetryDisabled, lastUpdated)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Constraint == "deployment_configuration_pkey" {
				// The row we were trying to insert already exists.
				// Don't treat this as an error.
				err = nil
			}

		}
	}
	return err
}
