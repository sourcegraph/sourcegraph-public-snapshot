BEGIN;

-- This migration is an extension of the combination of
-- migrations/frontend/1528395714_worker_num_failures.up.sql
-- and
-- migrations/frontend/1528395721_lsif_deleted.up.sql

-- We're introducing a new "state" to the workers, so we need to extend the two database enums that include the states:

--   lsif_index_state
--   lsif_upload_state

-- Unfortunately we can't add a value to a enum within a transaction, so we
-- have to make an entirely new enum and transfer all refrences to the old enum
-- to the new one. Hence the verbosity here.

-- Drop dependent views/indexes
DROP VIEW lsif_dumps_with_repository_name;
DROP VIEW lsif_indexes_with_repository_name;
DROP VIEW lsif_uploads_with_repository_name;
DROP VIEW lsif_dumps;
DROP INDEX lsif_uploads_repository_id_commit_root_indexer;

-- Create new temp enums

CREATE TYPE lsif_upload_state_temp AS ENUM(
    'uploading',
    'queued',
    'processing',
    'completed',
    'errored',
    'deleted',
    'failed' -- This is the new field
);

CREATE TYPE lsif_index_state_temp AS ENUM(
    'queued',
    'processing',
    'completed',
    'errored',
    'failed' -- This is the new field
);
--
-- Update type of state column that use the enums
ALTER TABLE lsif_uploads
    ALTER COLUMN state DROP DEFAULT,
    ALTER COLUMN state TYPE lsif_upload_state_temp USING state::text::lsif_upload_state_temp,
    ALTER COLUMN state SET DEFAULT 'queued';

ALTER TABLE lsif_indexes
    ALTER COLUMN state DROP DEFAULT,
    ALTER COLUMN state TYPE lsif_index_state_temp USING state::text::lsif_index_state_temp,
    ALTER COLUMN state SET DEFAULT 'queued';

-- Switch enum names
DROP TYPE lsif_upload_state;
ALTER TYPE lsif_upload_state_temp RENAME TO lsif_upload_state;

DROP TYPE lsif_index_state;
ALTER TYPE lsif_index_state_temp RENAME TO lsif_index_state;

-- Recreate views/indexes
CREATE UNIQUE INDEX lsif_uploads_repository_id_commit_root_indexer ON lsif_uploads(repository_id, "commit", root, indexer) WHERE state = 'completed'::lsif_upload_state;
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

CREATE VIEW lsif_dumps_with_repository_name AS
    SELECT u.*, r.name as repository_name FROM lsif_dumps u
    JOIN repo r ON r.id = u.repository_id
    WHERE r.deleted_at IS NULL;

CREATE VIEW lsif_uploads_with_repository_name AS
    SELECT u.*, r.name as repository_name FROM lsif_uploads u
    JOIN repo r ON r.id = u.repository_id
    WHERE r.deleted_at IS NULL;

CREATE VIEW lsif_indexes_with_repository_name AS
    SELECT u.*, r.name as repository_name FROM lsif_indexes u
    JOIN repo r ON r.id = u.repository_id
    WHERE r.deleted_at IS NULL;

COMMIT;

