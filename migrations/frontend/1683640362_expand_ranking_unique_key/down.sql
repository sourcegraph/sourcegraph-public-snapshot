CREATE UNIQUE INDEX IF NOT EXISTS codeintel_path_ranks_repository_id ON codeintel_path_ranks(repository_id);
DROP INDEX IF EXISTS codeintel_path_ranks_graph_key_repository_id;
ALTER TABLE codeintel_path_ranks ALTER COLUMN graph_key DROP NOT NULL;
