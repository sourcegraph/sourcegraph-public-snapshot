CREATE TABLE IF NOT EXISTS codeowners_owners (
    id SERIAL NOT NULL PRIMARY KEY,
    handle TEXT NULL,
    email TEXT NULL
);

CREATE INDEX IF NOT EXISTS codeowners_owners_handle_email ON codeowners_owners USING btree (handle, email);

-- Q: This can be just a column in repo_paths,
--    but then potentially needs different updated_at.
--    What's better?
CREATE TABLE IF NOT EXISTS codeowners_path_stats (
    file_path_id INTEGER NOT NULL PRIMARY KEY REFERENCES repo_paths(id),
    tree_owned_files_count INTEGER NOT NULL,
    last_updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS codeowners_individual_stats (
    file_path_id INTEGER NOT NULL REFERENCES repo_paths(id),
    owner_id INTEGER NOT NULL REFERENCES codeowners_owners(id),
    tree_owned_files_count INTEGER NOT NULL,
    last_updated_at TIMESTAMP NOT NULL
);

ALTER TABLE IF EXISTS repo_paths
ADD COLUMN IF NOT EXISTS tree_files_count INTEGER NULL;
