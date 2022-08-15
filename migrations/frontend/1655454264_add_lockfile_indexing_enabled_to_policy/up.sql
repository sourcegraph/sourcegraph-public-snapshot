ALTER TABLE lsif_configuration_policies
  ADD COLUMN IF NOT EXISTS lockfile_indexing_enabled boolean NOT NULL DEFAULT false;

COMMENT ON COLUMN lsif_configuration_policies.lockfile_indexing_enabled IS 'Whether to index the lockfiles in the repositories matched by this policy';
