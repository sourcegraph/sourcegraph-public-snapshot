DROP INDEX IF EXISTS repo_names_repo_id_name_unique_idx;

ALTER TABLE IF EXISTS repo_names
    DROP COLUMN IF EXISTS repo_id;

CREATE UNIQUE INDEX IF NOT EXISTS repo_names_name_unique_idx
    ON repo_names (name);
