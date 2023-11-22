ALTER TABLE ONLY sub_repo_permissions
    ADD COLUMN IF NOT EXISTS path_includes text[],
    ADD COLUMN IF NOT EXISTS path_excludes text[];
