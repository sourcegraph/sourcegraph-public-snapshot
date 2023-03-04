ALTER TABLE codeintel_ranking_definitions ADD COLUMN IF NOT EXISTS repository TEXT NOT NULL;
ALTER TABLE codeintel_ranking_path_counts_inputs DROP COLUMN IF EXISTS repository_id;
ALTER TABLE codeintel_ranking_path_counts_inputs ADD COLUMN IF NOT EXISTS repository TEXT NOT NULL;
