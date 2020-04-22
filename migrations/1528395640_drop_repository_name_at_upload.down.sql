BEGIN;

-- Drop view dependent on old column
DROP VIEW lsif_dumps;

-- Restore column
ALTER TABLE lsif_uploads ADD repository_name_at_upload TEXT;
UPDATE lsif_uploads SET repository_name_at_upload = 'unknown';
ALTER TABLE lsif_uploads ALTER repository_name_at_upload SET NOT NULL;

-- Recreate view with new column names
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

COMMIT;
