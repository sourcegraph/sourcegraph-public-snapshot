TRUNCATE codeintel_ranking_definitions CASCADE;
TRUNCATE codeintel_ranking_references CASCADE;
TRUNCATE codeintel_initial_path_ranks CASCADE;

ALTER TABLE codeintel_ranking_definitions DROP COLUMN IF EXISTS upload_id;
ALTER TABLE codeintel_ranking_references DROP COLUMN IF EXISTS upload_id;
ALTER TABLE codeintel_initial_path_ranks DROP COLUMN IF EXISTS upload_id;

ALTER TABLE codeintel_ranking_definitions ADD COLUMN IF NOT EXISTS exported_upload_id INTEGER NOT NULL REFERENCES codeintel_ranking_exports(id) ON DELETE CASCADE;
ALTER TABLE codeintel_ranking_references ADD COLUMN IF NOT EXISTS exported_upload_id INTEGER NOT NULL REFERENCES codeintel_ranking_exports(id) ON DELETE CASCADE;
ALTER TABLE codeintel_initial_path_ranks ADD COLUMN IF NOT EXISTS exported_upload_id INTEGER NOT NULL REFERENCES codeintel_ranking_exports(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS codeintel_ranking_definitions_exported_upload_id ON codeintel_ranking_definitions(exported_upload_id);
CREATE INDEX IF NOT EXISTS codeintel_ranking_references_exported_upload_id ON codeintel_ranking_references(exported_upload_id);
CREATE INDEX IF NOT EXISTS codeintel_initial_path_ranks_exported_upload_id ON codeintel_initial_path_ranks(exported_upload_id);
CREATE INDEX IF NOT EXISTS codeintel_ranking_definitions_graph_key_symbol_search ON codeintel_ranking_definitions(graph_key, symbol_name, exported_upload_id, document_path);
