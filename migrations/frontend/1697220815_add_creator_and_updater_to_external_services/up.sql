ALTER TABLE external_services ADD COLUMN IF NOT EXISTS creator_id integer, ADD COLUMN IF NOT EXISTS last_updater_id integer;
ALTER TABLE ONLY external_services
    DROP CONSTRAINT IF EXISTS external_services_creator_id_fkey,
    ADD CONSTRAINT external_services_creator_id_fkey
         FOREIGN KEY (creator_id)
            REFERENCES users(id)
            ON DELETE SET NULL DEFERRABLE;
ALTER TABLE ONLY external_services
    DROP CONSTRAINT IF EXISTS external_services_last_updater_id_fkey,
    ADD CONSTRAINT external_services_last_updater_id_fkey
         FOREIGN KEY (last_updater_id)
            REFERENCES users(id)
            ON DELETE SET NULL DEFERRABLE;
