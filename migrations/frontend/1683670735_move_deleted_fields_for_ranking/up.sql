ALTER TABLE codeintel_ranking_definitions DROP COLUMN IF EXISTS last_scanned_at;
ALTER TABLE codeintel_ranking_definitions DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE codeintel_ranking_references DROP COLUMN IF EXISTS last_scanned_at;
ALTER TABLE codeintel_ranking_references DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE codeintel_initial_path_ranks DROP COLUMN IF EXISTS last_scanned_at;
ALTER TABLE codeintel_initial_path_ranks DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE codeintel_ranking_exports ADD COLUMN IF NOT EXISTS last_scanned_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS codeintel_ranking_exports_graph_key_last_scanned_at ON codeintel_ranking_exports(graph_key, last_scanned_at NULLS FIRST, id);
ALTER TABLE codeintel_ranking_exports ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;
