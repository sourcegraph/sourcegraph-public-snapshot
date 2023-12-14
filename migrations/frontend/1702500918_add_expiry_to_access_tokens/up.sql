ALTER TABLE access_tokens ADD COLUMN IF NOT EXISTS expires_at timestamp with time zone;
