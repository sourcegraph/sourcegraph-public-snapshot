ALTER TABLE IF EXISTS notebooks
    DROP COLUMN IF EXISTS namespace_user_id,
    DROP COLUMN IF EXISTS namespace_org_id,
    DROP COLUMN IF EXISTS updater_user_id,
    DROP CONSTRAINT IF EXISTS notebooks_has_max_1_namespace;

DROP INDEX IF EXISTS notebooks_namespace_user_id_idx;

DROP INDEX IF EXISTS notebooks_namespace_org_id_idx;
