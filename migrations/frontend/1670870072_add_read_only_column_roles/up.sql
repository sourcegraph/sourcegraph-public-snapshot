ALTER TABLE roles
    ADD COLUMN IF NOT EXISTS readonly BOOLEAN DEFAULT FALSE;

COMMENT ON COLUMN roles.readonly IS 'This is used to indicate whether a role is read-only or can be modified.';

UPDATE roles SET readonly = FALSE;

ALTER TABLE roles
    ALTER COLUMN readonly SET NOT NULL;
