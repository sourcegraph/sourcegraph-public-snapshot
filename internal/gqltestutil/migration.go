package gqltestutil

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var MigrationPollInterval = time.Second

// PollMigration will invoke the given function periodically with the current progress of the
// given migration. The loop will break once the function returns true or the given context
// is canceled.
func (c *Client) PollMigration(ctx context.Context, id string, f func(float64) bool) error {
	for {
		progress, err := c.GetMigrationProgress(id)
		if err != nil {
			return err
		}
		if f(progress) {
			return nil
		}

		select {
		case <-time.After(MigrationPollInterval):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Client) GetMigrationProgress(id string) (float64, error) {
	const query = `
		query GetMigrationStatus {
			outOfBandMigrations {
				id
				progress
			}
		}
	`

	var envelope struct {
		Data struct {
			OutOfBandMigrations []struct {
				ID       string
				Progress float64
			}
		}
	}
	if err := c.GraphQL("", query, nil, &envelope); err != nil {
		return 0, errors.Wrap(err, "request GraphQL")
	}

	for _, migration := range envelope.Data.OutOfBandMigrations {
		if migration.ID == id {
			return migration.Progress, nil
		}
	}

	return 0, errors.Newf("unknown oobmigration %q", id)
}

func (c *Client) SetMigrationDirection(id string, up bool) error {
	const query = `
		mutation SetMigrationDirection($id: ID!, $applyReverse: Boolean!) {
			setMigrationDirection(id: $id, applyReverse: $applyReverse) {
				alwaysNil
			}
		}
	`

	variables := map[string]any{
		"id":           id,
		"applyReverse": !up,
	}
	if err := c.GraphQL("", query, variables, nil); err != nil {
		return errors.Wrap(err, "request GraphQL")
	}

	return nil
}
