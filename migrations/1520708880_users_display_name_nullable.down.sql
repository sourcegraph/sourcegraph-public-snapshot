UPDATE users SET display_name='' WHERE display_name IS NULL;
ALTER TABLE users ALTER COLUMN display_name SET NOT NULL;
