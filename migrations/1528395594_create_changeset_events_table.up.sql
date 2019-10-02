BEGIN;

CREATE TABLE changeset_events (
  id bigserial PRIMARY KEY,
  changeset_id bigint NOT NULL REFERENCES changesets(id)
    ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE,
  kind text NOT NULL CHECK (kind != ''),
  source text NOT NULL CHECK (source != ''),
  key text NOT NULL CHECK (key != '') UNIQUE,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  metadata jsonb NOT NULL DEFAULT '{}'
    CHECK (jsonb_typeof(metadata) = 'object')
);

CREATE INDEX ON changeset_events(changeset_id);

COMMIT;
