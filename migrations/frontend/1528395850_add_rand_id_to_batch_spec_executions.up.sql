BEGIN;

ALTER TABLE IF EXISTS batch_spec_executions ADD COLUMN IF NOT EXISTS rand_id text NOT NULL;

CREATE INDEX batch_spec_executions_rand_id ON batch_spec_executions USING btree (rand_id);

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
