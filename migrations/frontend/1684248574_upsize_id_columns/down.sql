ALTER TABLE codeintel_ranking_progress
    ADD COLUMN new_max_definition_id INTEGER,
    ADD COLUMN new_max_reference_id INTEGER,
    ADD COLUMN new_max_path_id INTEGER;

UPDATE codeintel_ranking_progress
SET
    new_max_definition_id = max_definition_id,
    new_max_reference_id = max_reference_id,
    new_max_path_id = max_path_id;

ALTER TABLE codeintel_ranking_progress
    DROP COLUMN max_definition_id,
    DROP COLUMN max_reference_id,
    DROP COLUMN max_path_id,
    ALTER COLUMN new_max_definition_id SET NOT NULL,
    ALTER COLUMN new_max_reference_id SET NOT NULL,
    ALTER COLUMN new_max_path_id SET NOT NULL;

ALTER TABLE codeintel_ranking_progress RENAME new_max_definition_id TO max_definition_id;
ALTER TABLE codeintel_ranking_progress RENAME new_max_reference_id TO max_reference_id;
ALTER TABLE codeintel_ranking_progress RENAME new_max_path_id TO max_path_id;
