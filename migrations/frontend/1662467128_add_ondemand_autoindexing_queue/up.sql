CREATE TABLE IF NOT EXISTS codeintel_autoindex_queue (
    id SERIAL PRIMARY KEY,
    repository_id int NOT NULL,
    rev text NOT NULL,
    queued_at timestamptz NOT NULL DEFAULT NOW(),
    processed_at timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS codeintel_autoindex_queue_repository_id_commit ON codeintel_autoindex_queue(repository_id, rev);
