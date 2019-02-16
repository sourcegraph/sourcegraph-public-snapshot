-- Create a temporary non unique index that makes finding
-- duplicates by (external_service_id, external_id) fast.
CREATE INDEX repo_external_service_temp_index
ON repo (external_service_id, external_id);

-- Remove all old duplicate repos that have an external_id set and
-- update discussion threads foreign key constraints to the newer id.
WITH duplicates AS (
  DELETE FROM repo r1 USING repo r2
  WHERE r1.external_id IS NOT NULL AND r2.external_id IS NOT NULL
  AND r1.external_service_id = r2.external_service_id
  AND r1.external_id = r2.external_id
  AND r1.id < r2.id
  RETURNING r1.id AS id, r2.id AS new_id
)
UPDATE discussion_threads_target_repo
SET repo_id = duplicates.new_id
FROM duplicates
WHERE repo_id = duplicates.id;

-- Remove the temporary index.
DROP INDEX repo_external_service_temp_index;

-- Now that we have duplicates removed, create a unique composite
-- index that ensures no more duplicates are introduced.
CREATE UNIQUE INDEX repo_external_service_unique_idx
ON repo (external_service_type, external_service_id, external_id)
WHERE external_service_type IS NOT NULL
AND external_service_id IS NOT NULL
AND external_id IS NOT NULL;
