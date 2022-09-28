CREATE TABLE IF NOT EXISTS sampled_repo_names
(
    id      SERIAL
        CONSTRAINT sampled_repo_names_pk PRIMARY KEY,
    repo_id INT  NOT NULL,
    name    TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS sampled_repo_names_repo_name_unique_idx on sampled_repo_names(repo_id, name);
CREATE INDEX IF NOT EXISTS sampled_repo_names_trgm_idx ON sampled_repo_names USING gin (name gin_trgm_ops);
