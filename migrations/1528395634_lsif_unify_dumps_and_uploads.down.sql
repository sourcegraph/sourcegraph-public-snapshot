-- Note: `commit` is a reserved word, so it's quoted.

BEGIN;

-- Drop shim view
DROP VIEW lsif_dumps;

-- Restore duplicate column
ALTER TABLE lsif_uploads ADD processed_at TIMESTAMP WITH TIME ZONE;

-- Remove new default from uploaded_at
ALTER TABLE lsif_uploads ALTER COLUMN uploaded_at DROP DEFAULT;

-- Populate duplicate column before removing source
UPDATE lsif_uploads SET processed_at = finished_at;

-- Rename lsif_uploads to lsif_dumps
ALTER INDEX lsif_uploads_pkey RENAME TO lsif_dumps_pkey;
ALTER INDEX lsif_uploads_repository_commit_root RENAME TO lsif_dumps_repository_commit_root;
ALTER INDEX lsif_uploads_uploaded_at RENAME TO lsif_dumps_uploaded_at;
ALTER INDEX lsif_uploads_visible_repository_commit RENAME TO lsif_dumps_visible_repository_commit;
ALTER TABLE lsif_uploads RENAME CONSTRAINT lsif_uploads_commit_valid_chars TO lsif_dumps_commit_valid_chars;
ALTER TABLE lsif_uploads RENAME CONSTRAINT lsif_uploads_repository_check TO lsif_dumps_repository_check;
ALTER TABLE lsif_uploads RENAME TO lsif_dumps;

-- Drop new index
DROP INDEX lsif_uploads_state;

-- Restore uploads table
CREATE TABLE lsif_uploads (
    id BIGSERIAL PRIMARY KEY,
    repository TEXT NOT NULL,
    "commit" TEXT NOT NULL,
    root TEXT NOT NULL,
    filename TEXT NOT NULL,
    state lsif_upload_state NOT NULL DEFAULT 'queued',
    failure_summary TEXT,
    failure_stacktrace TEXT,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    tracing_context TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS lsif_uploads_state ON lsif_uploads(state);
CREATE INDEX IF NOT EXISTS lsif_uploads_uploaded_at ON lsif_uploads(uploaded_at);

-- Move all non-completed uploads back to the upload stable
INSERT INTO lsif_uploads (
    repository, "commit", root,
    filename, state, failure_summary, failure_stacktrace,
    uploaded_at, started_at, finished_at, tracing_context
) SELECT
    u.repository, u."commit", u.root,
    u.filename, u.state, u.failure_summary, u.failure_stacktrace,
    u.uploaded_at, u.started_at, u.finished_at, u.tracing_context
FROM lsif_dumps u;

-- Remove extra rows from dumps table
DELETE FROM lsif_dumps WHERE state != 'completed';

-- Remove new columns
ALTER TABLE lsif_dumps
    DROP filename,
    DROP state,
    DROP failure_summary,
    DROP failure_stacktrace,
    DROP started_at,
    DROP finished_at,
    DROP tracing_context;

-- Restore original unique constraint
ALTER TABLE lsif_dumps ADD CONSTRAINT lsif_dumps_repository_commit_root UNIQUE (repository, commit, root);

COMMIT;
