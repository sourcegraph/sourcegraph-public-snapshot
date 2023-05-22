AlTER TABLE codeintel_ranking_definitions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
AlTER TABLE codeintel_ranking_references ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
AlTER TABLE codeintel_initial_path_ranks ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
