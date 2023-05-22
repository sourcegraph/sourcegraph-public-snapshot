ALTER TABLE codeintel_ranking_progress DROP COLUMN IF EXISTS max_export_id;

ALTER TABLE codeintel_ranking_progress ADD COLUMN IF NOT EXISTS max_definition_id BIGINT;
ALTER TABLE codeintel_ranking_progress ADD COLUMN IF NOT EXISTS max_reference_id BIGINT;
ALTER TABLE codeintel_ranking_progress ADD COLUMN IF NOT EXISTS max_path_id BIGINT;

UPDATE codeintel_ranking_progress SET
    max_definition_id = 0,
    max_reference_id = 0,
    max_path_id = 0;

ALTER TABLE codeintel_ranking_progress ALTER COLUMN max_definition_id SET NOT NULL;
ALTER TABLE codeintel_ranking_progress ALTER COLUMN max_reference_id SET NOT NULL;
ALTER TABLE codeintel_ranking_progress ALTER COLUMN max_path_id SET NOT NULL;
