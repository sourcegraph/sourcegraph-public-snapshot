ALTER TABLE codeintel_path_ranks ADD COLUMN IF NOT EXISTS num_paths INT;
ALTER TABLE codeintel_path_ranks ADD COLUMN IF NOT EXISTS min_reference_count INT;
ALTER TABLE codeintel_path_ranks ADD COLUMN IF NOT EXISTS mean_reference_count FLOAT;
ALTER TABLE codeintel_path_ranks ADD COLUMN IF NOT EXISTS max_reference_count INT;
