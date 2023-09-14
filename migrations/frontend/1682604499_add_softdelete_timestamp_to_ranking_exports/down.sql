AlTER TABLE codeintel_ranking_definitions DROP COLUMN IF EXISTS deleted_at;
AlTER TABLE codeintel_ranking_references DROP COLUMN IF EXISTS deleted_at;
AlTER TABLE codeintel_initial_path_ranks DROP COLUMN IF EXISTS deleted_at;
