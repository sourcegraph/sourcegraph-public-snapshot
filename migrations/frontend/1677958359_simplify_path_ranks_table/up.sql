TRUNCATE codeintel_path_ranks;

-- Add new primary key
ALTER TABLE codeintel_path_ranks ADD COLUMN IF NOT EXISTS id BIGSERIAL;
ALTER TABLE codeintel_path_ranks DROP CONSTRAINT IF EXISTS codeintel_path_ranks_pkey;
ALTER TABLE codeintel_path_ranks ADD PRIMARY KEY (id);

-- Drop precision
DROP INDEX IF EXISTS codeintel_path_ranks_repository_id_precision;
ALTER TABLE codeintel_path_ranks DROP COLUMN IF EXISTS precision;

-- Add new unique index to replace precision
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_path_ranks_repository_id ON codeintel_path_ranks(repository_id);
