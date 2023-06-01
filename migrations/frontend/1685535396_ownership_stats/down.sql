ALTER TABLE IF EXISTS repo_paths DROP COLUMN IF EXISTS deep_file_count;
DROP INDEX IF EXISTS codeowners_stats_file_owner;
DROP TABLE IF EXISTS codeowners_stats;
