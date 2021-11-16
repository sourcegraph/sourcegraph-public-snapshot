BEGIN;

ALTER TABLE cm_monitors
	ALTER COLUMN namespace_user_id DROP NOT NULL;
COMMIT;
