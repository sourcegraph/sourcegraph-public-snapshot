ALTER TABLE codeintel_ranking_progress DROP COLUMN IF EXISTS max_definition_id;
ALTER TABLE codeintel_ranking_progress DROP COLUMN IF EXISTS max_reference_id;
ALTER TABLE codeintel_ranking_progress DROP COLUMN IF EXISTS max_path_id;

ALTER TABLE codeintel_ranking_progress ADD COLUMN IF NOT EXISTS max_export_id BIGINT;
UPDATE codeintel_ranking_progress SET max_export_id = 0;
ALTER TABLE codeintel_ranking_progress ALTER COLUMN max_export_id SET NOT NULL;
