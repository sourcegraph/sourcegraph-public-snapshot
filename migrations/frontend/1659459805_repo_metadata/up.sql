-- Perform migration here.

CREATE TABLE IF NOT EXISTS repo_metadata (
    repo_id INTEGER NOT NULL REFERENCES repo(id) ON DELETE CASCADE,
    key CITEXT NOT NULL,
    value CITEXT,
    PRIMARY KEY (repo_id, key, value)
);
