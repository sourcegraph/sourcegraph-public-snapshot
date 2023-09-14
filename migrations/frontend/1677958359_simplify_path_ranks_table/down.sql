TRUNCATE codeintel_path_ranks;

-- Drop new column/primary key
ALTER TABLE codeintel_path_ranks DROP COLUMN IF EXISTS id;

-- Add back old column
ALTER TABLE codeintel_path_ranks ADD COLUMN IF NOT EXISTS precision float NOT NULL;

-- Revert index changes
DROP INDEX IF EXISTS codeintel_path_ranks_repository_id;
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_path_ranks_repository_id_precision ON codeintel_path_ranks(repository_id, precision);
