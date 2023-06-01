CREATE TABLE IF NOT EXISTS codeowners_stats (
    file_path_id INTEGER NOT NULL REFERENCES repo_paths(id) ON DELETE CASCADE DEFERRABLE,
    codeowners_id INTEGER NOT NULL REFERENCES commit_authors(id),
    deep_file_count INTEGER,
    PRIMARY KEY (file_path_id, codeowners_id)
);
-- Need to pick better names - these are effectively ownerless stats
CREATE TABLE IF NOT EXISTS codeowners_aggregate_stats (
    file_path_id INTEGER NOT NULL PRIMARY KEY REFERENCES repo_paths(id) ON DELETE CASCADE DEFERRABLE,
    deep_file_count INTEGER
);
ALTER TABLE IF EXISTS repo_paths
ADD COLUMN IF NOT EXISTS deep_file_count INTEGER NULL;
