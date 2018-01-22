ALTER TABLE repo ALTER COLUMN created_at DROP NOT NULL;
ALTER TABLE repo ALTER COLUMN created_at DROP DEFAULT;
ALTER TABLE repo ADD COLUMN blocked boolean;
ALTER TABLE repo ADD COLUMN "owner" citext;
ALTER TABLE repo ADD COLUMN "name" citext;