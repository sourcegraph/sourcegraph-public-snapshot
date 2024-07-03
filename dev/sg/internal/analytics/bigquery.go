package analytics

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/sourcegraph/log"
	"google.golang.org/api/option"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/background"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

var (
	bq               *bigquery.Client
	stopSendingStuff = make(chan struct{})
)

func BackgroundInitBigQueryConnection(ctx context.Context, secretsStore *secrets.Store) {
	background.Run(ctx, func(ctx context.Context, backgroundOutput *std.Output) {
		token, err := secretsStore.GetExternal(ctx, secrets.ExternalSecret{
			Project: secrets.LocalDevProject,
			Name:    "SG_BIGQUERY_TOKEN",
		})
		if err != nil {
			backgroundOutput.WriteWarningf("failed to fetch BigQuery token for analytics upload", log.Error(err))
			return
		}

		bq, err = bigquery.NewClient(ctx, "sourcegraph-local-dev", option.WithCredentialsJSON([]byte(token)))
		if err != nil {
			backgroundOutput.WriteWarningf("failed to create BigQuery client for analytics", log.Error(err))
			return
		}

		// interrupt.
		go sendStuffInBackground()
	})
}

func sendStuffInBackground() {
	for {
		select {
		case <-stopSendingStuff:
			return
		}
	}

	ctx := context.Background()
	sessionID, commit, rollback, err := transact(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			// shadows top-level err
			if err := rollback(); err != nil {
				bq.Warn("failed to rollback transaction", log.Error(err))
			}
			return
		}
		if err := commit(); err != nil {
			bq.Warn("failed to commit transaction", log.Error(err))
		}
	}()

	for i, query := range SQLInserts(bq.dataset, data...) {
		q := bq.Query(query.Text)
		q.ConnectionProperties = append(q.ConnectionProperties, &bigquery.ConnectionProperty{
			Key:   "session_id",
			Value: sessionID,
		})
		// q.Parameters = toParams(query.Params)
		// Location must match that of the dataset(s) referenced in the query.
		q.Location = "US"

		if err := func() error {
			job, err := q.Run(ctx)
			if err != nil {
				return err
			}
			status, err := job.Wait(ctx)
			if err != nil {
				return err
			}
			if err := status.Err(); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			return fmt.Errorf("error performing bigquery bulk insert: %w", err)
		}
	}
}

func transact(ctx context.Context) (session string, commit, rollback func() error, _ error) {
	q := bq.Query("SELECT 1;")
	q.CreateSession = true
	q.Location = "US"
	job, err := q.Run(ctx)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to start bigquery session: starting query: %w", err)
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to start bigquery session: awaiting query: %w", err)
	}
	if err := status.Err(); err != nil {
		return "", nil, nil, fmt.Errorf("failed to start bigquery session: running query: %w", err)
	}

	session = job.LastStatus().Statistics.SessionInfo.SessionID

	q = bq.Query("BEGIN TRANSACTION")
	q.ConnectionProperties = append(q.ConnectionProperties, &bigquery.ConnectionProperty{
		Key:   "session_id",
		Value: session,
	})
	q.Location = "US"
	job, err = q.Run(ctx)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to start bigquery transaction: starting query: %w", err)
	}
	status, err = job.Wait(ctx)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to start bigquery transaction: awaiting query: %w", err)
	}
	if err := status.Err(); err != nil {
		return "", nil, nil, fmt.Errorf("failed to start bigquery transaction: running query: %w", err)
	}

	quitSession := func() {
		q = bq.Query("CALL BQ.ABORT_SESSION()")
		q.ConnectionProperties = append(q.ConnectionProperties, &bigquery.ConnectionProperty{
			Key:   "session_id",
			Value: session,
		})
		q.Location = "US"
		job, err = q.Run(ctx)
		if err != nil {
			bq.Warn("failed to quit bigquery session: starting query", log.Error(err))
			return
		}
		status, err = job.Wait(ctx)
		if err != nil {
			bq.Warn("failed to quit bigquery session: awaiting query", log.Error(err))
			return
		}
		if err := status.Err(); err != nil {
			bq.Warn("failed to quit bigquery session: running query", log.Error(err))
			return
		}
	}

	return session, func() error {
			defer quitSession()

			q = bq.Query("COMMIT TRANSACTION")
			q.ConnectionProperties = append(q.ConnectionProperties, &bigquery.ConnectionProperty{
				Key:   "session_id",
				Value: session,
			})
			q.Location = "US"
			job, err = q.Run(ctx)
			if err != nil {
				return fmt.Errorf("failed to commit bigquery transaction: starting query: %w", err)
			}
			status, err = job.Wait(ctx)
			if err != nil {
				return fmt.Errorf("failed to commit bigquery transaction: awaiting query: %w", err)
			}
			if err := status.Err(); err != nil {
				return fmt.Errorf("failed to commit bigquery transaction: running query: %w", err)
			}
			return nil
		}, func() error {
			defer quitSession()

			q = bq.Query("ROLLBACK TRANSACTION")
			q.ConnectionProperties = append(q.ConnectionProperties, &bigquery.ConnectionProperty{
				Key:   "session_id",
				Value: session,
			})
			q.Location = "US"
			job, err = q.Run(ctx)
			if err != nil {
				return fmt.Errorf("failed to rollback bigquery transaction: starting query: %w", err)
			}
			status, err = job.Wait(ctx)
			if err != nil {
				return fmt.Errorf("failed to rollback bigquery transaction: awaiting query: %w", err)
			}
			if err := status.Err(); err != nil {
				return fmt.Errorf("failed to rollback bigquery transaction: running query: %w", err)
			}
			return nil
		}, nil
}
