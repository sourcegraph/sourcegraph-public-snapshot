CREATE INDEX CONCURRENTLY IF NOT EXISTS changesets_changeset_specs ON changesets (current_spec_id, previous_spec_id);
