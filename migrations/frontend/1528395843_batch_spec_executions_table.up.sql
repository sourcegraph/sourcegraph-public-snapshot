-- +++
-- parent: 1528395842
-- +++

BEGIN;

CREATE TABLE IF NOT EXISTS batch_spec_executions (
  id              BIGSERIAL PRIMARY KEY,
  state           TEXT DEFAULT 'queued',
  failure_message TEXT,
  started_at      TIMESTAMP WITH TIME ZONE,
  finished_at     TIMESTAMP WITH TIME ZONE,
  process_after   TIMESTAMP WITH TIME ZONE,
  num_resets      INTEGER NOT NULL DEFAULT 0,
  num_failures    INTEGER NOT NULL DEFAULT 0,
  execution_logs  JSON[],
  worker_hostname TEXT NOT NULL DEFAULT '',

  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

  batch_spec TEXT NOT NULL,
  batch_spec_id integer REFERENCES batch_specs(id)
);

COMMIT;
