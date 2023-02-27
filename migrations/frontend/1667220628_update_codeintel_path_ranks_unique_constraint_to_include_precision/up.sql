DROP INDEX IF EXISTS codeintel_path_ranks_repository_id;

CREATE UNIQUE INDEX IF NOT EXISTS codeintel_path_ranks_repository_id_precision ON codeintel_path_ranks (repository_id, precision);
