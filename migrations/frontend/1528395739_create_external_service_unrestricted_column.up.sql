BEGIN;

ALTER TABLE external_services ADD COLUMN IF NOT EXISTS unrestricted BOOLEAN;
UPDATE external_services SET unrestricted = FALSE;
ALTER TABLE external_services ALTER COLUMN unrestricted SET DEFAULT FALSE;
ALTER TABLE external_services ALTER COLUMN unrestricted SET NOT NULL;

COMMIT;
