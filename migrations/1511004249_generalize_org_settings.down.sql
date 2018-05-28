ALTER TABLE settings RENAME TO org_settings;
ALTER TABLE org_settings RENAME CONSTRAINT settings_references_orgs TO org_settings_references_orgs;
-- ALTER TABLE org_settings RENAME CONSTRAINT settings_author_user_id_fkey TO org_settings_references_users;
ALTER SEQUENCE settings_id_seq RENAME TO org_settings_id_seq;
ALTER INDEX settings_pkey RENAME TO org_settings_pkey;
ALTER TABLE org_settings ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE org_settings DROP CONSTRAINT has_subject;
