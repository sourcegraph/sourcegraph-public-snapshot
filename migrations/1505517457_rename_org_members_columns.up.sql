ALTER TABLE org_members RENAME COLUMN user_name TO username;
ALTER TABLE org_members RENAME COLUMN user_email TO email;
ALTER TABLE org_members ADD COLUMN display_name TEXT;
ALTER TABLE org_members ADD COLUMN avatar_url TEXT;