CREATE UNIQUE INDEX IF NOT EXISTS codeintel_ranking_exports_graph_key_upload_id ON codeintel_ranking_exports(graph_key, upload_id);

DROP INDEX IF EXISTS codeintel_ranking_exports_upload_id_graph_key;
