BEGIN;

CREATE TABLE changeset_events (
  id bigserial PRIMARY KEY,
  changeset_id bigint NOT NULL REFERENCES changesets(id)
    ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE,
  kind text NOT NULL CHECK (kind != ''),
  source text NOT NULL CHECK (source != ''),
  key text NOT NULL CHECK (key != ''),
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  metadata jsonb NOT NULL DEFAULT '{}'
    CHECK (jsonb_typeof(metadata) = 'object')
);

ALTER TABLE changeset_events
ADD CONSTRAINT changeset_events_changeset_id_kind_key_unique
UNIQUE (changeset_id, kind, key);

COMMIT;
