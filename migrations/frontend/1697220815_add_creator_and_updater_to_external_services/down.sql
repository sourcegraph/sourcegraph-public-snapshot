ALTER TABLE external_services DROP COLUMN IF EXISTS creator_id, DROP COLUMN IF EXISTS last_updater_id;
ALTER TABLE external_services DROP CONSTRAINT IF EXISTS external_services_creator_id_fkey;
ALTER TABLE external_services DROP CONSTRAINT IF EXISTS external_services_last_updater_id_fkey;
