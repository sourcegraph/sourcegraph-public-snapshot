BEGIN;

DROP VIEW lsif_dumps;

ALTER TABLE lsif_uploads DROP COLUMN failure_stacktrace;
ALTER TABLE lsif_indexes DROP COLUMN failure_stacktrace;
ALTER TABLE lsif_uploads RENAME COLUMN failure_summary TO failure_message;
ALTER TABLE lsif_indexes RENAME COLUMN failure_summary TO failure_message;

-- Recreate view with new columns
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

COMMIT;
