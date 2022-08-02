-- Perform migration here.

CREATE TABLE IF NOT EXISTS repo_metadata (
    repo_id INTEGER NOT NULL REFERENCES repo(id),
    key citext NOT NULL,
    value citext,
    PRIMARY KEY (repo_id, citext, value)
);
