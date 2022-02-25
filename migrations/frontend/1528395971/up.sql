ALTER TABLE IF EXISTS notebooks
    ADD COLUMN IF NOT EXISTS namespace_user_id integer REFERENCES users(id) ON DELETE SET NULL DEFERRABLE,
    ADD COLUMN IF NOT EXISTS namespace_org_id integer REFERENCES orgs(id) ON DELETE SET NULL DEFERRABLE,
    ADD COLUMN IF NOT EXISTS updater_user_id integer REFERENCES users(id) ON DELETE SET NULL DEFERRABLE,
    DROP CONSTRAINT IF EXISTS notebooks_has_max_1_namespace,
    ADD CONSTRAINT notebooks_has_max_1_namespace CHECK ((namespace_user_id IS NULL AND namespace_org_id IS NULL) OR ((namespace_user_id IS NULL) <> (namespace_org_id IS NULL)));

CREATE INDEX IF NOT EXISTS notebooks_namespace_user_id_idx ON notebooks USING btree (namespace_user_id);

CREATE INDEX IF NOT EXISTS notebooks_namespace_org_id_idx ON notebooks USING btree (namespace_org_id);

UPDATE notebooks SET namespace_user_id = creator_user_id;
