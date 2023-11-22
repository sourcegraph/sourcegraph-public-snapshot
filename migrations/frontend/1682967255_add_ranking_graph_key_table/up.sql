CREATE TABLE IF NOT EXISTS codeintel_ranking_progress (
    id                          BIGSERIAL PRIMARY KEY,
    graph_key                   TEXT NOT NULL UNIQUE,
    max_definition_id           INTEGER NOT NULL,
    max_reference_id            INTEGER NOT NULL,
    max_path_id                 INTEGER NOT NULL,
    mappers_started_at          TIMESTAMP WITH TIME ZONE NOT NULL,
    mapper_completed_at         TIMESTAMP WITH TIME ZONE,
    seed_mapper_completed_at    TIMESTAMP WITH TIME ZONE,
    reducer_started_at          TIMESTAMP WITH TIME ZONE,
    reducer_completed_at        TIMESTAMP WITH TIME ZONE
);
