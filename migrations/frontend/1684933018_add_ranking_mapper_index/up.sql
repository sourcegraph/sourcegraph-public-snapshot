CREATE INDEX CONCURRENTLY IF NOT EXISTS codeintel_ranking_exports_graph_key_deleted_at_id ON codeintel_ranking_exports(graph_key, deleted_at DESC NULLS FIRST, id);
