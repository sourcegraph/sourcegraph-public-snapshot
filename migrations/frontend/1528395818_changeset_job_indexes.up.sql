BEGIN;

CREATE INDEX IF NOT EXISTS changeset_jobs_state_idx ON changeset_jobs USING BTREE(state);
CREATE INDEX IF NOT EXISTS changeset_jobs_group_idx ON changeset_jobs USING BTREE(bulk_group);

COMMIT;
