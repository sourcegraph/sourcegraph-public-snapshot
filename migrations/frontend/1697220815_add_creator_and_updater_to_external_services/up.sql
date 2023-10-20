ALTER TABLE external_services ADD COLUMN IF NOT EXISTS creator_id integer, ADD COLUMN IF NOT EXISTS last_updater_id integer;
