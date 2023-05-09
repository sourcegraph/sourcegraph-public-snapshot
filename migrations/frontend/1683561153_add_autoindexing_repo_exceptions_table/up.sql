CREATE TABLE IF NOT EXISTS codeintel_autoindexing_exceptions(
    id                  SERIAL PRIMARY KEY,
    repository_id       INTEGER NOT NULL UNIQUE,
    disable_scheduling  BOOLEAN NOT NULL DEFAULT FALSE,
    disable_inference   BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (repository_id) REFERENCES repo(id) ON DELETE CASCADE
);
