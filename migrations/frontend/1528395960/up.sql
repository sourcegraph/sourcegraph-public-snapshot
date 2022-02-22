ALTER TABLE batch_spec_workspaces ADD COLUMN IF NOT EXISTS skipped_steps integer[] NOT NULL DEFAULT '{}';
