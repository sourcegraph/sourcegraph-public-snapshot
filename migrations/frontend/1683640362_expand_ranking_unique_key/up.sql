DELETE FROM codeintel_path_ranks WHERE graph_key IS NULL;
ALTER TABLE codeintel_path_ranks ALTER COLUMN graph_key SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_path_ranks_graph_key_repository_id ON codeintel_path_ranks(graph_key, repository_id);
DROP INDEX IF EXISTS codeintel_path_ranks_repository_id;
