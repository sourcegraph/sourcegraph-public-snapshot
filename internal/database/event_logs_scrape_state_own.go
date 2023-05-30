package database

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type eventLogsScrapeStateStore struct {
	*basestore.Store
}

func EventLogsScrapeStateStoreWith(other basestore.ShareableStore) EventLogsScrapeStateStore {
	return &eventLogsScrapeStateStore{Store: basestore.NewWithHandle(other.Handle())}
}

type EventLogsScrapeStateStore interface {
	GetBookmark(ctx context.Context, signalName string) (int, error)
	UpdateBookmark(ctx context.Context, val int, signalName string) error
}

func (s *eventLogsScrapeStateStore) GetBookmark(ctx context.Context, signalName string) (int, error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	val, found, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
		`SELECT bookmark_id FROM event_logs_scrape_state_own 
				WHERE job_type = (select id from own_signal_configurations where name = %s) 
				ORDER BY id LIMIT 1`, signalName)))
	if err != nil {
		return 0, err
	}
	if !found {
		// generate a row and return the value
		return basestore.ScanInt(tx.QueryRow(ctx, sqlf.Sprintf(
			`INSERT INTO event_logs_scrape_state_own (bookmark_id, job_type) 
					SELECT MAX(id), (select id from own_signal_configurations where name = %s) 
					FROM event_logs RETURNING bookmark_id`, signalName)))
	}
	return val, err
}

func (s *eventLogsScrapeStateStore) UpdateBookmark(ctx context.Context, val int, signalName string) error {
	return s.Exec(ctx, sqlf.Sprintf(
		`UPDATE event_logs_scrape_state_own SET bookmark_id = %d 
				WHERE id = (SELECT id FROM event_logs_scrape_state_own WHERE job_type = (select id from own_signal_configurations where name = %s) 
				ORDER BY id LIMIT 1)`, val, signalName))
}
