-- Note: `commit` is a reserved word, so it's quoted.

BEGIN;

-- Add lsif_upload columns to lsif_dumps
ALTER TABLE lsif_dumps
    ADD filename TEXT,
    ADD state lsif_upload_state NOT NULL DEFAULT 'queued',
    ADD failure_summary TEXT,
    ADD failure_stacktrace TEXT,
    ADD started_at TIMESTAMP WITH TIME ZONE,
    ADD finished_at TIMESTAMP WITH TIME ZONE,
    ADD tracing_context TEXT;

-- Transfer default value from field
ALTER TABLE lsif_dumps ALTER COLUMN uploaded_at SET DEFAULT now();

-- Replace unique constraint with unique partial index
ALTER TABLE lsif_dumps DROP CONSTRAINT lsif_dumps_repository_commit_root;
CREATE UNIQUE INDEX lsif_dumps_repository_commit_root ON lsif_dumps (repository, commit, root) WHERE (state = 'completed');

-- Set dumb values for each 'legacy' dump
UPDATE lsif_dumps SET
    filename = '<unknown>',
    state = 'completed',
    started_at = processed_at,
    finished_at = processed_at,
    tracing_context = '{}';

-- Drop duplicate column
ALTER TABLE lsif_dumps DROP COLUMN processed_at;

-- Move all queued, in-progress, and failed uploads into the same table
INSERT INTO lsif_dumps (
    repository, "commit", root,
    filename, state, failure_summary, failure_stacktrace,
    uploaded_at, started_at, finished_at, tracing_context
) SELECT
    u.repository, u."commit", u.root,
    u.filename, u.state, u.failure_summary, u.failure_stacktrace,
    u.uploaded_at, u.started_at, u.finished_at, u.tracing_context
FROM lsif_uploads u WHERE state != 'completed';

-- Drop old table
DROP TABLE lsif_uploads;

-- Set NOT NULL constraints after population
ALTER TABLE lsif_dumps ALTER COLUMN filename SET NOT NULL;
ALTER TABLE lsif_dumps ALTER COLUMN tracing_context SET NOT NULL;

-- Rename lsif_dumps to lsif_uploads
ALTER TABLE lsif_dumps RENAME TO lsif_uploads;
ALTER TABLE lsif_uploads RENAME CONSTRAINT lsif_dumps_commit_valid_chars TO lsif_uploads_commit_valid_chars;
ALTER TABLE lsif_uploads RENAME CONSTRAINT lsif_dumps_repository_check TO lsif_uploads_repository_check;
ALTER INDEX lsif_dumps_pkey RENAME TO lsif_uploads_pkey;
ALTER INDEX lsif_dumps_repository_commit_root RENAME TO lsif_uploads_repository_commit_root;
ALTER INDEX lsif_dumps_uploaded_at RENAME TO lsif_uploads_uploaded_at;
ALTER INDEX lsif_dumps_visible_repository_commit RENAME TO lsif_uploads_visible_repository_commit;

-- Create missing index
CREATE INDEX lsif_uploads_state ON lsif_uploads(state);

-- Create a view into lsif_dumps to minimize code changes
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

COMMIT;
