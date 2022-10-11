ALTER TABLE webhooks
    DROP COLUMN IF EXISTS created_by,
    DROP COLUMN IF EXISTS updated_by;
