BEGIN;

ALTER TABLE external_services ADD COLUMN IF NOT EXISTS encryption_key_id text NOT NULL DEFAULT '';

COMMIT;
