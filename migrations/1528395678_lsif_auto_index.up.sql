BEGIN;

CREATE TABLE lsif_indexable_repositories (
    id SERIAL PRIMARY KEY,
    repository_id integer NOT NULL UNIQUE,
    search_count integer NOT NULL DEFAULT 0,
    precise_count integer NOT NULL DEFAULT 0,
    last_index_enqueued_at timestamp with time zone
);

CREATE TYPE lsif_index_state AS ENUM (
    'queued',
    'processing',
    'completed',
    'errored'
);

CREATE TABLE lsif_indexes (
    id SERIAL PRIMARY KEY,
    commit text NOT NULL,
    queued_at timestamp with time zone DEFAULT now() NOT NULL,
    state lsif_index_state DEFAULT 'queued'::lsif_index_state NOT NULL,
    failure_summary text,
    failure_stacktrace text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    repository_id integer NOT NULL,
    CONSTRAINT lsif_uploads_commit_valid_chars CHECK ((commit ~ '^[a-z0-9]{40}$'::text))
);

COMMIT;
