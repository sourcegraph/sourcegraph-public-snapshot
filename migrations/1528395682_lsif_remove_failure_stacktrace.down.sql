BEGIN;

DROP VIEW lsif_dumps;

ALTER TABLE lsif_uploads ADD COLUMN failure_stacktrace text;
ALTER TABLE lsif_indexes ADD COLUMN failure_stacktrace text;
ALTER TABLE lsif_uploads RENAME COLUMN failure_message TO failure_summary;
ALTER TABLE lsif_indexes RENAME COLUMN failure_message TO failure_summary;

-- Recreate view with original columns
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

COMMIT;
