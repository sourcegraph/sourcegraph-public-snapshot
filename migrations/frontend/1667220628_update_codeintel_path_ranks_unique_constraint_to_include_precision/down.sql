DROP INDEX IF EXISTS codeintel_path_ranks_repository_id_precision;

CREATE UNIQUE INDEX IF NOT EXISTS codeintel_path_ranks_repository_id ON codeintel_path_ranks (repository_id);
