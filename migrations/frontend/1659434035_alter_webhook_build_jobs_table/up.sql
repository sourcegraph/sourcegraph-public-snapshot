ALTER TABLE IF EXISTS webhook_build_jobs
    ADD COLUMN IF NOT EXISTS org text,
    ADD COLUMN IF NOT EXISTS extsvc_id integer;
