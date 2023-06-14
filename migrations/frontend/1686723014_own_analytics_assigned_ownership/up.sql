ALTER TABLE IF EXISTS ownership_path_stats
ADD COLUMN IF NOT EXISTS tree_assigned_ownership_files_count INTEGER NULL,
    ADD COLUMN IF NOT EXISTS tree_any_ownership_files_count INTEGER NULL;