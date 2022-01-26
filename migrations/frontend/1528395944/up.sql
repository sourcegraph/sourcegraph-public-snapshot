-- +++
-- parent: 1528395943
-- +++

BEGIN;
	-- If code monitors were disabled by a manual step, make
	-- it possible to re-enable them by removing the constraint.
	-- If an admin wants to restore the previous enabled state
	-- from the backup table, they can run something like the following:
	--     UPDATE cm_monitors
	--     SET enabled = cm_monitors_enabled_backup.enabled
	--     FROM cm_monitors_enabled_backup
	--     WHERE cm_monitors.id = cm_monitors_enabled_backup.id;
	ALTER TABLE cm_monitors
	DROP CONSTRAINT IF EXISTS cm_monitors_cannot_be_enabled;
COMMIT;
