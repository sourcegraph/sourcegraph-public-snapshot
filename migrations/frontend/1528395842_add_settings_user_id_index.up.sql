-- +++
-- parent: 1528395841
-- +++

BEGIN;

CREATE INDEX IF NOT EXISTS settings_user_id_idx ON settings USING BTREE (user_id);

COMMIT;
