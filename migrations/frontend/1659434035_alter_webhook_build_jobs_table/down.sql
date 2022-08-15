ALTER TABLE IF EXISTS webhook_build_jobs
    DROP COLUMN IF EXISTS org,
    DROP COLUMN IF EXISTS extsvc_id;
