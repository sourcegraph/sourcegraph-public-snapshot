BEGIN;

ALTER TABLE lsif_configuration_policies ADD COLUMN repository_patterns TEXT[];
COMMENT ON COLUMN lsif_configuration_policies.repository_patterns IS 'The name pattern matching repositories to which this configuration policy applies. If absent, all repositories are matched.';

COMMIT;
