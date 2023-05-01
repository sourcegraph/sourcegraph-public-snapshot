ALTER TABLE codeintel_initial_path_ranks ALTER COLUMN document_path SET DEFAULT '';
ALTER TABLE codeintel_initial_path_ranks ADD COLUMN IF NOT EXISTS document_paths TEXT[] NOT NULL DEFAULT '{}';
