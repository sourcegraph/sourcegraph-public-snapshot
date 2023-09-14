ALTER TABLE codeintel_ranking_exports DROP COLUMN IF EXISTS last_scanned_at;
ALTER TABLE codeintel_ranking_exports DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE codeintel_ranking_definitions ADD COLUMN IF NOT EXISTS last_scanned_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS codeintel_ranking_definitions_graph_key_last_scanned_at ON codeintel_ranking_definitions(graph_key, last_scanned_at NULLS FIRST, id);
ALTER TABLE codeintel_ranking_definitions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE codeintel_ranking_references ADD COLUMN IF NOT EXISTS last_scanned_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS codeintel_ranking_references_graph_key_last_scanned_at ON codeintel_ranking_references(graph_key, last_scanned_at NULLS FIRST, id);
ALTER TABLE codeintel_ranking_references ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE codeintel_initial_path_ranks ADD COLUMN IF NOT EXISTS last_scanned_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS codeintel_initial_path_ranks_graph_key_last_scanned_at ON codeintel_initial_path_ranks(graph_key, last_scanned_at NULLS FIRST, id);
ALTER TABLE codeintel_initial_path_ranks ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

DROP INDEX codeintel_ranking_definitions_graph_key_last_scanned_at;
CREATE INDEX IF NOT EXISTS codeintel_ranking_definitions_graph_key_last_scanned_at_id ON codeintel_ranking_definitions USING btree (graph_key, last_scanned_at NULLS FIRST, id);
DROP INDEX codeintel_ranking_references_graph_key_last_scanned_at;
CREATE INDEX IF NOT EXISTS codeintel_ranking_references_graph_key_last_scanned_at_id ON codeintel_ranking_references USING btree (graph_key, last_scanned_at NULLS FIRST, id);
