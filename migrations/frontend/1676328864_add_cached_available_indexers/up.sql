CREATE TABLE IF NOT EXISTS cached_available_indexers (
    id SERIAL PRIMARY KEY,
    repository_id INTEGER NOT NULL,
    num_events INTEGER NOT NULL,
    available_indexers JSONB NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS cached_available_indexers_repository_id ON cached_available_indexers(repository_id);
CREATE INDEX IF NOT EXISTS cached_available_indexers_num_events ON cached_available_indexers(num_events DESC) WHERE available_indexers::text != '{}';
