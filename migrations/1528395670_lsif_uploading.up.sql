BEGIN;

-- Changes:
--   - add num_parts column to lsif_uploads
--   - add uploaded_parts column to lsif_uploads
--   - add 'uploading' state to lsif_upload_state enum
--
-- Unfortunatley we can't add a value to a enum within a transaction, so we have to make
-- an entirely new enum and transfer all refrences to the old enum to the new one. Hence
-- the verbosity here.

-- Drop view and index that depends on this type
DROP VIEW lsif_dumps;
DROP INDEX lsif_uploads_repository_id_commit_root_indexer;

-- Create new enum
CREATE TYPE lsif_upload_state_temp AS ENUM(
    'uploading',
    'queued',
    'processing',
    'completed',
    'errored'
);

-- The actual change
ALTER TABLE lsif_uploads
    ADD COLUMN num_parts int,
    ADD COLUMN uploaded_parts int[],
    ALTER COLUMN state DROP DEFAULT,
    ALTER COLUMN state TYPE lsif_upload_state_temp USING state::text::lsif_upload_state_temp,
    ALTER COLUMN state SET DEFAULT 'queued';

-- Backfill the new columns
UPDATE lsif_uploads SET num_parts = 1, uploaded_parts = '{0}';

-- Make them non-nullable
ALTER TABLE lsif_uploads
    ALTER COLUMN num_parts SET NOT NULL,
    ALTER COLUMN uploaded_parts SET NOT NULL;

-- Switch enum names
DROP TYPE lsif_upload_state;
ALTER TYPE lsif_upload_state_temp RENAME TO lsif_upload_state;

-- Restore index and view
CREATE UNIQUE INDEX lsif_uploads_repository_id_commit_root_indexer ON lsif_uploads(repository_id, "commit", root, indexer) WHERE state = 'completed'::lsif_upload_state;
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

COMMIT;
