TRUNCATE codeintel_path_ranks;

DROP INDEX IF EXISTS codeintel_path_ranks_repository_id_precision;
ALTER TABLE codeintel_path_ranks ADD COLUMN IF NOT EXISTS id BIGSERIAL PRIMARY KEY;
ALTER TABLE codeintel_path_ranks DROP COLUMN IF EXISTS precision;
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_path_ranks_repository_id ON codeintel_path_ranks(repository_id);
