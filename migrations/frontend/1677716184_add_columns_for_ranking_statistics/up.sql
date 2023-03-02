ALTER TABLE codeintel_path_ranks ADD COLUMN IF NOT EXISTS num_paths INT;
ALTER TABLE codeintel_path_ranks ADD COLUMN IF NOT EXISTS refcount_logsum FLOAT;
