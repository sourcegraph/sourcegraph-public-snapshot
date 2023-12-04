DROP TYPE IF EXISTS lsif_index_state;
DROP TYPE IF EXISTS lsif_upload_state;

CREATE TYPE lsif_index_state AS ENUM (
    'queued',
    'processing',
    'completed',
    'errored',
    'failed'
);

CREATE TYPE lsif_upload_state AS ENUM (
    'uploading',
    'queued',
    'processing',
    'completed',
    'errored',
    'deleted',
    'failed'
);
