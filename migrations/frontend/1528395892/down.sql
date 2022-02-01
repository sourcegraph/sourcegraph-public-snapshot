BEGIN;

CREATE TABLE IF NOT EXISTS batch_spec_executions (
  id              BIGSERIAL PRIMARY KEY,
  rand_id         TEXT NOT NULL,

  state           TEXT DEFAULT 'queued',
  failure_message TEXT,
  process_after   TIMESTAMP WITH TIME ZONE,
  started_at      TIMESTAMP WITH TIME ZONE,
  finished_at     TIMESTAMP WITH TIME ZONE,
  last_heartbeat_at TIMESTAMP WITH TIME ZONE,
  num_resets      INTEGER NOT NULL DEFAULT 0,
  num_failures    INTEGER NOT NULL DEFAULT 0,
  execution_logs  JSON[],
  worker_hostname TEXT NOT NULL DEFAULT '',
  cancel          BOOL DEFAULT FALSE,

  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

  batch_spec TEXT NOT NULL,
  batch_spec_id integer REFERENCES batch_specs(id) DEFERRABLE,

  user_id INTEGER REFERENCES users(id),
  namespace_org_id INTEGER REFERENCES orgs(id),
  namespace_user_id INTEGER REFERENCES users(id)
);

ALTER TABLE IF EXISTS batch_spec_executions ADD CONSTRAINT batch_spec_executions_has_1_namespace CHECK ((namespace_user_id IS NULL) <> (namespace_org_id IS NULL));
CREATE INDEX IF NOT EXISTS batch_spec_executions_rand_id ON batch_spec_executions USING btree (rand_id);

COMMIT;
