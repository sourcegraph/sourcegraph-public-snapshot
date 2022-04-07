ALTER TABLE batch_spec_workspaces ADD COLUMN IF NOT EXISTS steps JSONB DEFAULT '[]'::JSONB CHECK (jsonb_typeof(steps) = 'array'::TEXT);
ALTER TABLE batch_spec_workspaces ADD COLUMN IF NOT EXISTS skipped_steps INTEGER[] NOT NULL DEFAULT '{}'::INTEGER[];
