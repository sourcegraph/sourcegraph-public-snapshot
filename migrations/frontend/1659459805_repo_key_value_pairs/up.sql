-- Perform migration here.

CREATE TABLE IF NOT EXISTS repo_kvps (
    repo_id INTEGER NOT NULL REFERENCES repo(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT NULL,
    PRIMARY KEY (repo_id, key) INCLUDE (value)
);
