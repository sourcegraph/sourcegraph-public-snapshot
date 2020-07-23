BEGIN;

DROP VIEW lsif_dumps;

ALTER TABLE lsif_uploads DROP COLUMN process_after;
ALTER TABLE lsif_indexes DROP COLUMN process_after;

-- Recreate view with original columns
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

COMMIT;
