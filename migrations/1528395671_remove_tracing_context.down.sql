BEGIN;

-- Drop view and index that depends on this type
DROP VIEW lsif_dumps;

-- Add back the column
ALTER TABLE lsif_uploads ADD COLUMN tracing_context TEXT;
UPDATE lsif_uploads SET tracing_context = '{}';
ALTER TABLE lsif_uploads ALTER COLUMN tracing_context SET NOT NULL;

-- Restore the view
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

COMMIT;
