ALTER TABLE codeintel_ranking_definitions DROP COLUMN IF EXISTS repository;
ALTER TABLE codeintel_ranking_path_counts_inputs DROP COLUMN IF EXISTS repository;
ALTER TABLE codeintel_ranking_path_counts_inputs ADD COLUMN IF NOT EXISTS repository_id INTEGER NOT NULL;
