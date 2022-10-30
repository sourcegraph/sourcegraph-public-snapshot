CREATE
OR REPLACE AGGREGATE sg_jsonb_concat_agg(jsonb) (
    SFUNC = 'jsonb_concat',
    STYPE = jsonb,
    INITCOND = '{}'
);

ALTER TABLE
    codeintel_path_ranks
ALTER COLUMN
    payload TYPE jsonb USING payload :: jsonb;

CREATE TABLE IF NOT EXISTS codeintel_path_rank_inputs(
    id BIGSERIAL PRIMARY KEY,
    graph_key text NOT NULL,
    input_filename text NOT NULL,
    repository_name text NOT NULL,
    payload jsonb NOT NULL,
    processed boolean NOT NULL DEFAULT false,
    UNIQUE (graph_key, input_filename, repository_name)
);

CREATE INDEX IF NOT EXISTS codeintel_path_rank_graph_key_id_repository_name_processed ON codeintel_path_rank_inputs(graph_key, id, repository_name)
WHERE
    NOT processed;

COMMENT ON TABLE codeintel_path_rank_inputs IS 'Sharded inputs from Spark jobs that will subsequently be written into `codeintel_path_ranks`.';
