BEGIN;

-- Create lookup table for repository pattern matching
CREATE TABLE IF NOT EXISTS lsif_configuration_policies_repository_pattern_lookup (
    policy_id INTEGER NOT NULL,
    repo_id INTEGER NOT NULL,
    PRIMARY KEY (policy_id, repo_id)
);

COMMENT ON TABLE lsif_configuration_policies_repository_pattern_lookup IS 'A lookup table to get all the repository patterns by repository id that apply to a configuration policy.';
COMMENT ON COLUMN lsif_configuration_policies_repository_pattern_lookup.policy_id IS 'The policy identifier associated with the repository.';
COMMENT ON COLUMN lsif_configuration_policies_repository_pattern_lookup.repo_id IS 'The repository identifier associated with the policy.';

-- Add glob pattern column
ALTER TABLE lsif_configuration_policies ADD COLUMN repository_patterns TEXT[];
COMMENT ON COLUMN lsif_configuration_policies.repository_patterns IS 'The name pattern matching repositories to which this configuration policy applies. If absent, all repositories are matched.';

-- Add column to determine the last update of the associated records in lsif_configuration_policies_repository_pattern_lookup
ALTER TABLE lsif_configuration_policies ADD COLUMN last_resolved_at TIMESTAMP WITH TIME ZONE DEFAULT NULL;

COMMIT;
