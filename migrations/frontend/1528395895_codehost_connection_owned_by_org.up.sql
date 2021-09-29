-- Adds support for organization owning a codehost connection
BEGIN;

ALTER TABLE IF EXISTS external_services ADD COLUMN IF NOT EXISTS namespace_org_id integer REFERENCES orgs(id) ON DELETE CASCADE DEFERRABLE;
ALTER TABLE IF EXISTS external_services ADD CONSTRAINT external_services_max_1_namespace CHECK ((namespace_user_id IS NULL AND namespace_org_id IS NULL) OR ((namespace_user_id IS NULL) <> (namespace_org_id IS NULL)));
CREATE INDEX external_services_namespace_org_id_idx ON external_services USING btree (namespace_org_id);

END;