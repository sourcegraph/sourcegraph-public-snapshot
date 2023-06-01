CREATE TABLE IF NOT EXISTS codeowners_stats (
    file_path_id INTEGER NOT NULL REFERENCES repo_paths(id) ON DELETE CASCADE DEFERRABLE,
    codeowners_id INTEGER NOT NULL REFERENCES commit_authors(id),
    deep_file_count INTEGER,
    PRIMARY KEY (file_path_id, codeowners_id)
);
ALTER TABLE IF EXISTS repo_paths
ADD COLUMN IF NOT EXISTS deep_file_count INTEGER NULL;
