BEGIN;

CREATE TABLE IF NOT EXISTS versions (
  service text PRIMARY KEY,
  version text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);

COMMIT;
