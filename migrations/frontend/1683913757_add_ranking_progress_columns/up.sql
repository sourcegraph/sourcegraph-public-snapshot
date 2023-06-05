ALTER TABLE codeintel_ranking_progress ADD COLUMN IF NOT EXISTS num_path_records_total INT;
ALTER TABLE codeintel_ranking_progress ADD COLUMN IF NOT EXISTS num_reference_records_total INT;
ALTER TABLE codeintel_ranking_progress ADD COLUMN IF NOT EXISTS num_count_records_total INT;
ALTER TABLE codeintel_ranking_progress ADD COLUMN IF NOT EXISTS num_path_records_processed INT;
ALTER TABLE codeintel_ranking_progress ADD COLUMN IF NOT EXISTS num_reference_records_processed INT;
ALTER TABLE codeintel_ranking_progress ADD COLUMN IF NOT EXISTS num_count_records_processed INT;
