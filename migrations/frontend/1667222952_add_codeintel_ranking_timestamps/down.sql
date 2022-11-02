DROP TRIGGER IF EXISTS update_codeintel_path_ranks_updated_at ON codeintel_path_ranks;
DROP FUNCTION IF EXISTS update_codeintel_path_ranks_updated_at_column;
DROP INDEX IF EXISTS codeintel_path_ranks_updated_at;
ALTER TABLE codeintel_path_ranks DROP COLUMN IF EXISTS updated_at;
