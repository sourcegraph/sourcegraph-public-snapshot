CREATE TABLE IF NOT EXISTS codeintel_ranking_exports (
    upload_id integer NOT NULL,
    graph_key text NOT NULL,
    locked_at timestamp with time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE
    codeintel_ranking_exports
ADD
    PRIMARY KEY (upload_id, graph_key);

ALTER TABLE
    codeintel_ranking_exports
ADD
    FOREIGN KEY (upload_id) REFERENCES lsif_uploads(id) ON DELETE CASCADE;
