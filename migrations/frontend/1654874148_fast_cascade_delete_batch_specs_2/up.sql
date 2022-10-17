-- Create index for foreign key reverse lookups. This is required for cascading deletes.
CREATE INDEX CONCURRENTLY IF NOT EXISTS batch_spec_workspaces_batch_spec_id ON batch_spec_workspaces (batch_spec_id);
