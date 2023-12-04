-- Rename default constraint name to something we can control
ALTER TABLE codeintel_path_ranks DROP CONSTRAINT IF EXISTS codeintel_path_ranks_repository_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_path_ranks_repository_id ON codeintel_path_ranks (repository_id);
