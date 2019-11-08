BEGIN;

ALTER TABLE changesets DROP COLUMN campaign_job_id;
ALTER TABLE changesets DROP COLUMN error;

-- Delete entries with NULL external_id
DELETE FROM changesets WHERE external_id IS NULL;
ALTER TABLE changesets ALTER COLUMN external_id SET NOT NULL;
-- Drop `external_id IS NULL OR external_id != ''` check
ALTER TABLE changesets DROP CONSTRAINT changesets_external_id_check;
-- Recreate check without `IS NULL`
ALTER TABLE changesets ADD CHECK (external_id != '');

-- New constraints are dropped by `DROP COLUMN above new constraint
-- Add old constraint
ALTER TABLE changesets
ADD CONSTRAINT changesets_repo_external_id_unique UNIQUE (repo_id, external_id);

COMMIT;
