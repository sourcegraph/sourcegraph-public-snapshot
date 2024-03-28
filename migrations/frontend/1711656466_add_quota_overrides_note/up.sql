-- Add a note field to allow for context when modifying
-- the completions_quota and code_completions_quota columns.
ALTER TABLE users ADD COLUMN IF NOT EXISTS completions_quota_note text DEFAULT '';
