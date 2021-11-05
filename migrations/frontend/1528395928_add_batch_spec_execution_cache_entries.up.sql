BEGIN;

CREATE TABLE IF NOT EXISTS batch_spec_execution_cache_entries (
  id           BIGSERIAL PRIMARY KEY,

  key          TEXT NOT NULL,
  value        TEXT NOT NULL,

  version      INTEGER NOT NULL,

  last_used_at TIMESTAMP WITH TIME ZONE,
  created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMIT;
