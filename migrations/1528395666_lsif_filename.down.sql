BEGIN;

-- Drop view dependent on column
DROP VIEW lsif_dumps;

-- Add back dummy filename column
ALTER TABLE lsif_uploads ADD COLUMN filename TEXT;
UPDATE lsif_uploads SET filename = '';
ALTER TABLE lsif_uploads ALTER COLUMN filename SET NOT NULL;

-- Recreate view with renamed columns
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

COMMIT;
