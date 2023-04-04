ALTER TABLE codeintel_initial_path_ranks DROP COLUMN IF EXISTS document_paths;
ALTER TABLE codeintel_initial_path_ranks ALTER COLUMN document_path DROP DEFAULT;
