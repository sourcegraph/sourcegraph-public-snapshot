BEGIN;
ALTER TABLE cm_queries
    ADD COLUMN IF NOT EXISTS next_run timestamptz default now(),
    ADD COLUMN IF NOT EXISTS latest_result timestamptz;
COMMIT;
