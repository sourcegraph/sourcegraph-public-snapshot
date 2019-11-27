-- Note: `commit` is a reserved word, so it's quoted.

BEGIN;

CREATE TYPE lsif_upload_state AS ENUM (
    'queued',
    'processing',
    'completed',
    'errored'
);

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
    completed_or_errored_at TIMESTAMP WITH TIME ZONE,
    tracing_context TEXT NOT NULL
);

COMMIT;
