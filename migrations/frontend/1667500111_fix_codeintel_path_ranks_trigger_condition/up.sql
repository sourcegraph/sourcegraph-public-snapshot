DROP TRIGGER IF EXISTS update_codeintel_path_ranks_updated_at ON codeintel_path_ranks;
CREATE TRIGGER update_codeintel_path_ranks_updated_at BEFORE UPDATE ON codeintel_path_ranks FOR EACH ROW WHEN (NEW IS DISTINCT FROM OLD) EXECUTE PROCEDURE update_codeintel_path_ranks_updated_at_column();
