ALTER TABLE external_service_repos ADD COLUMN IF NOT EXISTS user_id integer REFERENCES users(id) ON DELETE CASCADE DEFERRABLE;
ALTER TABLE external_service_repos ADD COLUMN IF NOT EXISTS org_id integer REFERENCES orgs(id) ON DELETE CASCADE;
ALTER TABLE external_services ADD COLUMN IF NOT EXISTS namespace_user_id integer;
ALTER TABLE external_services DROP CONSTRAINT IF EXISTS external_services_namepspace_user_id_fkey;
ALTER TABLE external_services ADD CONSTRAINT external_services_namepspace_user_id_fkey FOREIGN KEY(namespace_user_id) REFERENCES users(id) ON DELETE CASCADE DEFERRABLE;
ALTER TABLE external_services ADD COLUMN IF NOT EXISTS namespace_org_id integer REFERENCES orgs(id) ON DELETE CASCADE DEFERRABLE;
CREATE UNIQUE INDEX IF NOT EXISTS external_services_unique_kind_org_id ON external_services (kind, namespace_org_id) WHERE (deleted_at IS NULL AND namespace_user_id IS NULL AND namespace_org_id IS NOT NULL);
CREATE UNIQUE INDEX IF NOT EXISTS external_services_unique_kind_user_id ON external_services (kind, namespace_user_id) WHERE (deleted_at IS NULL AND namespace_org_id IS NULL AND namespace_user_id IS NOT NULL);
CREATE INDEX IF NOT EXISTS external_services_namespace_org_id_idx ON external_services USING btree (namespace_org_id);
CREATE INDEX IF NOT EXISTS external_services_namespace_user_id_idx ON external_services USING btree (namespace_user_id);
CREATE INDEX IF NOT EXISTS external_service_user_repos_idx ON external_service_repos USING btree (user_id, repo_id) WHERE (user_id IS NOT NULL);
ALTER TABLE external_services DROP CONSTRAINT IF EXISTS external_services_max_1_namespace;
ALTER TABLE external_services ADD CONSTRAINT external_services_max_1_namespace CHECK ((((namespace_user_id IS NULL) AND (namespace_org_id IS NULL)) OR ((namespace_user_id IS NULL) <> (namespace_org_id IS NULL))));
CREATE INDEX IF NOT EXISTS external_service_repos_org_id_idx ON external_service_repos USING btree (org_id) WHERE org_id IS NOT NULL;
DROP TRIGGER IF EXISTS trig_soft_delete_user_reference_on_external_service ON users;
DROP FUNCTION IF EXISTS soft_delete_user_reference_on_external_service;
CREATE FUNCTION soft_delete_user_reference_on_external_service() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- If a user is soft-deleted, delete every row that references that user
    IF (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL) THEN
        UPDATE external_services
        SET deleted_at = NOW()
        WHERE namespace_user_id = OLD.id;
    END IF;

    RETURN OLD;
END;
$$;
CREATE TRIGGER trig_soft_delete_user_reference_on_external_service AFTER UPDATE OF deleted_at ON users FOR EACH ROW EXECUTE FUNCTION soft_delete_user_reference_on_external_service();
