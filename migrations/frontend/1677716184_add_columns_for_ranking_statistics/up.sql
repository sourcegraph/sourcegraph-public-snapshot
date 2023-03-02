ALTER TABLE codeintel_path_ranks ADD COLUMN IF NOT EXISTS num_paths INT;
ALTER TABLE codeintel_path_ranks ADD COLUMN IF NOT EXISTS min_reference_count INT;
ALTER TABLE codeintel_path_ranks ADD COLUMN IF NOT EXISTS max_reference_count INT;
ALTER TABLE codeintel_path_ranks ADD COLUMN IF NOT EXISTS sum_reference_count INT;
