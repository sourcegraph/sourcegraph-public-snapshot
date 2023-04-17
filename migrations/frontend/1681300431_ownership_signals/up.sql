CREATE TABLE IF NOT EXISTS repo_paths (
    id SERIAL PRIMARY KEY,
    repo_id INTEGER NOT NULL REFERENCES repo(id) ON DELETE CASCADE DEFERRABLE,
    absolute_path TEXT NOT NULL,
    is_dir BOOLEAN NOT NULL,
    parent_id INTEGER NULL REFERENCES repo_paths(id)
);

COMMENT ON COLUMN repo_paths.absolute_path
IS 'Absolute path does not start or end with forward slash. Example: "a/b/c". Root directory is empty path "".';

CREATE UNIQUE INDEX IF NOT EXISTS repo_paths_index_absolute_path
ON repo_paths
USING btree (repo_id, absolute_path);

CREATE TABLE IF NOT EXISTS commit_authors (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    name TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS commit_authors_email_name
ON commit_authors
USING btree (email, name);

CREATE TABLE IF NOT EXISTS own_signal_recent_contribution (
    id SERIAL PRIMARY KEY,
    commit_author_id INTEGER NOT NULL REFERENCES commit_authors(id),
    changed_file_path_id INTEGER NOT NULL REFERENCES repo_paths(id),
    commit_timestamp TIMESTAMP NOT NULL,
    commit_id_hash INTEGER NOT NULL
);

COMMENT ON TABLE own_signal_recent_contribution
IS 'One entry per file changed in every commit that classifies as a contribution signal.';

CREATE TABLE IF NOT EXISTS own_aggregate_recent_contribution (
    id SERIAL PRIMARY KEY,
    commit_author_id INTEGER NOT NULL REFERENCES commit_authors(id),
    changed_file_path_id INTEGER NOT NULL REFERENCES repo_paths(id),
    contributions_count INTEGER DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS own_aggregate_recent_contribution_author_file
ON own_aggregate_recent_contribution
USING btree (commit_author_id, changed_file_path_id);
