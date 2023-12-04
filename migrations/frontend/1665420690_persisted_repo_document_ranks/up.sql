CREATE TABLE IF NOT EXISTS codeintel_path_ranks (
    repository_id integer NOT NULL UNIQUE,
    payload text NOT NULL
);
