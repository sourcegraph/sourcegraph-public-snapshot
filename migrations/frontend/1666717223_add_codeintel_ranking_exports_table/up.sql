CREATE TABLE IF NOT EXISTS codeintel_ranking_exports (
    upload_id integer NOT NULL,
    graph_key text NOT NULL,
    locked_at timestamp with time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY (upload_id, graph_key),
    FOREIGN KEY (upload_id) REFERENCES lsif_uploads(id) ON DELETE CASCADE
);
