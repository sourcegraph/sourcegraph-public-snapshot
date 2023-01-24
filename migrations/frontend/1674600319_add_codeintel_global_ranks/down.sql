-- Undo the changes made in the up migration
-- DROP TRIGGER IF EXISTS update_codeintel_global_ranks_updated_at ON codeintel_global_ranks;
-- DROP FUNCTION IF EXISTS update_codeintel_global_ranks_updated_at_column;
DROP TABLE IF EXISTS codeintel_global_ranks CASCADE;
