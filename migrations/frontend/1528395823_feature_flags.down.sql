BEGIN;

DROP TABLE IF EXISTS feature_flag_overrides;

DROP TABLE IF EXISTS feature_flags;

DROP TYPE feature_flag_type;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
