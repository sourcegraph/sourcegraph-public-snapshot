ALTER TABLE IF EXISTS repo_paths DROP COLUMN IF EXISTS tree_files_count,
    DROP COLUMN IF EXISTS tree_files_counts_updated_at;

DROP TABLE IF EXISTS codeowners_individual_stats;

DROP TABLE IF EXISTS ownership_path_stats;

DROP INDEX IF EXISTS codeowners_owners_reference;
DROP TABLE IF EXISTS codeowners_owners;
