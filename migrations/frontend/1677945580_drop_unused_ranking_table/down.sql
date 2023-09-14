CREATE TABLE IF NOT EXISTS codeintel_path_rank_inputs (
    id BIGSERIAL PRIMARY KEY,
    graph_key text NOT NULL,
    input_filename text NOT NULL,
    repository_name text NOT NULL,
    payload jsonb NOT NULL,
    processed boolean DEFAULT false NOT NULL,
    "precision" double precision NOT NULL,
    UNIQUE (graph_key, input_filename, repository_name)
);

CREATE INDEX IF NOT EXISTS codeintel_path_rank_inputs_graph_key_repository_name_id_process ON codeintel_path_rank_inputs USING btree (graph_key, repository_name, id) WHERE (NOT processed);
