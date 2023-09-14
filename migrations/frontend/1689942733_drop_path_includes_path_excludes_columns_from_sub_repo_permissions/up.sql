ALTER TABLE ONLY sub_repo_permissions
    DROP COLUMN IF EXISTS path_includes,
    DROP COLUMN IF EXISTS path_excludes;
