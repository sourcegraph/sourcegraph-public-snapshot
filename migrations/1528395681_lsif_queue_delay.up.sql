BEGIN;

DROP VIEW lsif_dumps;

ALTER TABLE lsif_uploads ADD COLUMN process_after timestamp with time zone;
ALTER TABLE lsif_indexes ADD COLUMN process_after timestamp with time zone;

-- Recreate view with new columns
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

COMMIT;
