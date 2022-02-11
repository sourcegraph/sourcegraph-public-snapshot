BEGIN;

CREATE INDEX IF NOT EXISTS settings_global_id ON settings (id DESC) WHERE user_id IS NULL AND org_id IS NULL;

COMMIT;
