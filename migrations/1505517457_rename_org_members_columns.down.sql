ALTER TABLE org_members RENAME COLUMN username TO user_name;
ALTER TABLE org_members RENAME COLUMN email TO user_email;
ALTER TABLE org_members DROP COLUMN display_name;
ALTER TABLE org_members DROP COLUMN avatar_url;