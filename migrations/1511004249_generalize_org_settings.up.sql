ALTER TABLE org_settings RENAME TO settings;
ALTER TABLE settings RENAME CONSTRAINT org_settings_references_orgs TO settings_references_orgs;
ALTER TABLE settings RENAME CONSTRAINT org_settings_references_users TO settings_references_users;
ALTER SEQUENCE org_settings_id_seq RENAME TO settings_id_seq;
ALTER INDEX org_settings_pkey RENAME TO settings_pkey;

-- Make sure that there is a subject (what the settings apply to). The only
-- supported subject is an org now, but this supports adding more in the future.
ALTER TABLE settings ADD CONSTRAINT has_subject CHECK (org_id IS NOT NULL);
ALTER TABLE settings ALTER COLUMN org_id DROP NOT NULL;
