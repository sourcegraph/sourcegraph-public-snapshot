BEGIN;

-- Drop view dependent on column
DROP VIEW lsif_dumps;

-- Rename column
ALTER TABLE lsif_uploads RENAME payload_id TO filename;

-- Recreate view with new column names
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

COMMIT;
