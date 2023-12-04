ALTER TABLE codeintel_ranking_progress DROP COLUMN IF EXISTS num_path_records_total;
ALTER TABLE codeintel_ranking_progress DROP COLUMN IF EXISTS num_reference_records_total;
ALTER TABLE codeintel_ranking_progress DROP COLUMN IF EXISTS num_count_records_total;
ALTER TABLE codeintel_ranking_progress DROP COLUMN IF EXISTS num_path_records_processed;
ALTER TABLE codeintel_ranking_progress DROP COLUMN IF EXISTS num_reference_records_processed;
ALTER TABLE codeintel_ranking_progress DROP COLUMN IF EXISTS num_count_records_processed;
