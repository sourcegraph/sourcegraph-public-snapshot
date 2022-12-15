CREATE UNIQUE INDEX IF NOT EXISTS codeintel_ranking_exports_upload_id_graph_key ON codeintel_ranking_exports(upload_id, graph_key);

DROP INDEX IF EXISTS codeintel_ranking_exports_graph_key_upload_id;
