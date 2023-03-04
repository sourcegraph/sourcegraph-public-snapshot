DROP INDEX IF EXISTS codeintel_path_ranks_repository_id;
ALTER TABLE codeintel_path_ranks DROP COLUMN IF EXISTS id;
ALTER TABLE codeintel_path_ranks ADD COLUMN IF NOT EXISTS precision float NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_path_ranks_repository_id_precision ON codeintel_path_ranks(repository_id, precision);
