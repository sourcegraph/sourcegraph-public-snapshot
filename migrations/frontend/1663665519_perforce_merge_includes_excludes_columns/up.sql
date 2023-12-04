ALTER TABLE ONLY sub_repo_permissions
    ADD COLUMN IF NOT EXISTS paths text[];

COMMENT ON COLUMN sub_repo_permissions.paths IS 'Paths that begin with a minus sign (-) are exclusion paths.';

UPDATE sub_repo_permissions
    SET paths = (ARRAY(SELECT CONCAT('/', path_include) FROM unnest(path_includes) as path_include) || ARRAY(SELECT CONCAT('-/', path_exclude) FROM unnest(path_excludes) as path_exclude));
