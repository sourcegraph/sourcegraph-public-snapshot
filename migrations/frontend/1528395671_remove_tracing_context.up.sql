BEGIN;

-- Drop view and index that depends on this type
DROP VIEW lsif_dumps;

-- Remove the column
ALTER TABLE lsif_uploads DROP COLUMN tracing_context;

-- Restore the view
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

COMMIT;
