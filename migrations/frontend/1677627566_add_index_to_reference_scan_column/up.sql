CREATE INDEX CONCURRENTLY IF NOT EXISTS codeintel_ranking_references_graph_key_last_scanned_at_id
ON codeintel_ranking_references(graph_key, last_scanned_at NULLS FIRST, id);
