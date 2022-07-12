-- Create index for foreign key reverse lookups. This is required for cascading deletes.
CREATE INDEX CONCURRENTLY IF NOT EXISTS changeset_specs_batch_spec_id ON changeset_specs (batch_spec_id);
