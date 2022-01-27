BEGIN;

ALTER TABLE batch_spec_workspaces ADD COLUMN IF NOT EXISTS steps JSONB DEFAULT '[]'::jsonb CHECK (jsonb_typeof(steps) = 'array'::text);
ALTER TABLE batch_spec_workspaces ADD COLUMN IF NOT EXISTS skipped_steps INTEGER[] NOT NULL DEFAULT '{}'::INTEGER{};

COMMIT;
