package database

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type SpongeLogStore interface {
	Save(context.Context, SpongeLog) error
	ByID(context.Context, uuid.UUID) (SpongeLog, error)
}

type logStore struct {
	*basestore.Store
}

type SpongeLog struct {
	ID          uuid.UUID
	Text        string
	Interpreter string
}

type spongeLogNotFoundErr struct{}

func (spongeLogNotFoundErr) Error() string  { return "sponge log not found" }
func (spongeLogNotFoundErr) NotFound() bool { return true }

var insertIgnoreLogFmtstr = `
	WITH existing (id) AS (
		SELECT i.id
		FROM sponge_log_interpreters AS i
		WHERE i.name = %s
	), inserted (id) AS (
		INSERT INTO sponge_log_interpreters (name)
		SELECT %s
		WHERE NOT EXISTS (
			SELECT j.id
			FROM sponge_log_interpreters AS j
			WHERE j.name = %s
		)
		RETURNING id
	)
	SELECT id FROM existing
	UNION ALL
	SELECT id FROM inserted
`

var upsertLogFmtstr = `
	INSERT INTO sponge_logs (id, log, interpreter_id)
	VALUES (%s, %s, NULLIF(%s, 0))
	ON CONFLICT (id) DO UPDATE SET
	log = EXCLUDED.log,
	interpreter_id = NULLIF(EXCLUDED.interpreter_id, 0)
`

func (s *logStore) Save(ctx context.Context, log SpongeLog) error {
	var interpreterID int
	if name := log.Interpreter; name != "" {
		q := sqlf.Sprintf(insertIgnoreLogFmtstr, name, name, name)
		if err := s.QueryRow(ctx, q).Scan(&interpreterID); err != nil {
			return err
		}
	}
	q := sqlf.Sprintf(upsertLogFmtstr, log.ID, log.Text, interpreterID)
	if err := s.Exec(ctx, q); err != nil {
		return err
	}
	return nil
}

var logByUUIDFmtstr = `
	SELECT l.id, l.log, COALESCE(i.name, '')
	FROM sponge_logs AS l
	LEFT JOIN sponge_log_interpreters AS i ON l.interpreter_id = i.id
	WHERE l.id = %s
`

func (s *logStore) ByID(ctx context.Context, id uuid.UUID) (SpongeLog, error) {
	var log SpongeLog
	q := sqlf.Sprintf(logByUUIDFmtstr, id)
	err := s.QueryRow(ctx, q).Scan(&log.ID, &log.Text, &log.Interpreter)
	if errors.Is(err, sql.ErrNoRows) {
		return log, spongeLogNotFoundErr{}
	}
	return log, err
}
