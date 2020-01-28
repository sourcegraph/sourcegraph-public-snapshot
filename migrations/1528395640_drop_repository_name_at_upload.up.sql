BEGIN;

-- Drop view dependent on old column
DROP VIEW lsif_dumps;

-- Drop column
ALTER TABLE lsif_uploads DROP IF EXISTS repository_name_at_upload;

-- Recreate view with new column names
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

COMMIT;
