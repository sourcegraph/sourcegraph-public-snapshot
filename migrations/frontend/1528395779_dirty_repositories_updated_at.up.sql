BEGIN;

ALTER TABLE lsif_dirty_repositories ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE;
COMMENT ON COLUMN lsif_dirty_repositories.updated_at IS 'The time the update_token value was last updated.';

COMMIT;
