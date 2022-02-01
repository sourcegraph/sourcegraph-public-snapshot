BEGIN;

ALTER TABLE lsif_uploads_visible_at_tip DROP COLUMN branch_or_tag_name;
ALTER TABLE lsif_uploads_visible_at_tip DROP COLUMN is_default_branch;

COMMIT;
