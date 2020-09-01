BEGIN;

ALTER TABLE changesets ADD COLUMN IF NOT EXISTS enqueued BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE changesets SET enqueued = TRUE where reconciler_state = 'queued';

COMMIT;
