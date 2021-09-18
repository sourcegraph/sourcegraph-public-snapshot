BEGIN;

ALTER TABLE lsif_uploads_visible_at_tip DROP COLUMN branch_or_tag_name;
ALTER TABLE lsif_uploads_visible_at_tip DROP COLUMN is_default_branch;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
