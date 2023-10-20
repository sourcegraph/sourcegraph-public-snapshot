-- Undo the changes made in the up migration
ALTER TABLE external_services DROP COLUMN IF EXISTS creator_id, DROP COLUMN IF EXISTS last_updater_id;
