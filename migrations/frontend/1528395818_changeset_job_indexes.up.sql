BEGIN;

-- Create an index on state, so the dbworker can fetch pending jobs faster.
CREATE INDEX IF NOT EXISTS changeset_jobs_state_idx ON changeset_jobs USING BTREE(state);
-- Create an index on the bulk_group column. We use this as sort of a second-line
-- primary key, because the bulk_group entity doesn't really exist.
CREATE INDEX IF NOT EXISTS changeset_jobs_group_idx ON changeset_jobs USING BTREE(bulk_group);

COMMIT;
