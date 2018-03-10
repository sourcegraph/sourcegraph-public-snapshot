ALTER TABLE users ALTER COLUMN display_name DROP NOT NULL;
UPDATE users SET display_name=NULL WHERE display_name='';
