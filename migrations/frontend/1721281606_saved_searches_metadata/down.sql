ALTER TABLE saved_searches DROP COLUMN IF EXISTS created_by;
ALTER TABLE saved_searches DROP COLUMN IF EXISTS updated_by;
ALTER TABLE saved_searches DROP COLUMN IF EXISTS draft;
ALTER TABLE saved_searches DROP COLUMN IF EXISTS visibility_secret;
