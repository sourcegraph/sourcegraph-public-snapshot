TRUNCATE codeintel_ranking_definitions CASCADE;
TRUNCATE codeintel_ranking_references CASCADE;
TRUNCATE codeintel_initial_path_ranks CASCADE;

ALTER TABLE codeintel_ranking_definitions DROP COLUMN IF EXISTS exported_upload_id;
ALTER TABLE codeintel_ranking_references DROP COLUMN IF EXISTS exported_upload_id;
ALTER TABLE codeintel_initial_path_ranks DROP COLUMN IF EXISTS exported_upload_id;

ALTER TABLE codeintel_ranking_definitions ADD COLUMN IF NOT EXISTS upload_id INTEGER NOT NULL;
ALTER TABLE codeintel_ranking_references ADD COLUMN IF NOT EXISTS upload_id INTEGER NOT NULL;
ALTER TABLE codeintel_initial_path_ranks ADD COLUMN IF NOT EXISTS upload_id INTEGER NOT NULL;

CREATE INDEX IF NOT EXISTS codeintel_ranking_definitions_graph_key_symbol_search ON codeintel_ranking_definitions(graph_key, symbol_name, upload_id, document_path);
CREATE INDEX IF NOT EXISTS codeintel_ranking_references_upload_id ON codeintel_ranking_references(upload_id);
CREATE INDEX IF NOT EXISTS codeintel_initial_path_ranks_upload_id ON codeintel_initial_path_ranks(upload_id);
CREATE INDEX IF NOT EXISTS codeintel_initial_path_upload_id ON codeintel_initial_path_ranks USING btree (upload_id);
DROP INDEX codeintel_initial_path_ranks_upload_id;
