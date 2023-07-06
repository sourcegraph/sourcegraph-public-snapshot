ALTER TABLE users ADD COLUMN IF NOT EXISTS completed_post_signup BOOLEAN DEFAULT FALSE;

-- TODO: Migrate users with verified emails to have this field set to true
