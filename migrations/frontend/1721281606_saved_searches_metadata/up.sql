ALTER TABLE saved_searches ADD COLUMN IF NOT EXISTS created_by integer REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE saved_searches ADD COLUMN IF NOT EXISTS updated_by integer REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE saved_searches ADD COLUMN IF NOT EXISTS draft boolean NOT NULL DEFAULT false;
ALTER TABLE saved_searches ADD COLUMN IF NOT EXISTS visibility_secret boolean NOT NULL DEFAULT true;
