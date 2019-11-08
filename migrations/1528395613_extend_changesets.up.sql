BEGIN;

ALTER TABLE changesets ADD COLUMN campaign_job_id bigint REFERENCES campaign_jobs(id)
    DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE changesets ALTER COLUMN external_id DROP NOT NULL;
-- Drop old `external_id != ''` check
ALTER TABLE changesets DROP CONSTRAINT changesets_external_id_check;
-- Recreate check but only if entry is non-NULL
ALTER TABLE changesets ADD CHECK (external_id IS NULL OR external_id != '');

-- Drop old `changesets_repo_external_id_unique` constraint
ALTER TABLE changesets DROP CONSTRAINT changesets_repo_external_id_unique;

-- Recreate constraint but including campaign_job_id
-- ALTER TABLE changesets
-- ADD CONSTRAINT changesets_unique UNIQUE (repo_id, external_id, campaign_job_id);

CREATE UNIQUE INDEX changesets_unique ON changesets (repo_id, external_id, campaign_job_id)
WHERE (external_id IS NOT NULL AND campaign_job_id IS NOT NULL);

ALTER TABLE changesets
ADD CONSTRAINT changesets_external_id_or_campaign_job_id_set
CHECK (external_id IS NOT NULL OR campaign_job_id IS NOT NULL);

ALTER TABLE changesets ADD COLUMN error TEXT;
CREATE INDEX changesets_error ON changesets(error);

COMMIT;
