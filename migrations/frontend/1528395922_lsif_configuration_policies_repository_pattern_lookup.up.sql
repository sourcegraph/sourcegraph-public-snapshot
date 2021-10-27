BEGIN;

-- Add new column for a timestamp where the lsif_configuration_policies_repository_pattern_lookup table was last updated
ALTER TABLE lsif_configuration_policies ADD COLUMN last_resolved_at TIMESTAMP WITH TIME ZONE DEFAULT NULL;

-- Create lookup table for repository pattern matching
CREATE TABLE IF NOT EXISTS lsif_configuration_policies_repository_pattern_lookup (
    policy_id INTEGER NOT NULL,
    repo_id INTEGER NOT NULL,
    PRIMARY KEY (policy_id, repo_id)
);

COMMENT ON TABLE lsif_configuration_policies_repository_pattern_lookup IS 'A lookup table to get all the repository patterns by repository id that apply to a configuration policy.';
COMMENT ON COLUMN lsif_configuration_policies_repository_pattern_lookup.policy_id IS 'The policy identifier associated with the repository.';
COMMENT ON COLUMN lsif_configuration_policies_repository_pattern_lookup.repo_id IS 'The repository identifier associated with the policy.';

COMMIT;
