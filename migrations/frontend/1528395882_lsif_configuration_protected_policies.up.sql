BEGIN;

ALTER TABLE lsif_configuration_policies ADD COLUMN protected boolean DEFAULT false;
UPDATE lsif_configuration_policies SET protected = false;
ALTER TABLE lsif_configuration_policies ALTER COLUMN protected SET NOT NULL;

COMMENT ON COLUMN lsif_configuration_policies.protected IS 'Whether or not this configuration policy is protected from modification of its data retention behavior (except for duration).';

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
