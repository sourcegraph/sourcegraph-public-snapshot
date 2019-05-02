BEGIN;

-- We make the URI column nullable. Undo this change.
UPDATE repo SET uri = name WHERE uri IS NULL;
ALTER TABLE repo ALTER COLUMN uri SET NOT NULL;

-- Matches the trigger part of 1528395556_.up.sql migration.
CREATE OR REPLACE FUNCTION set_repo_name() RETURNS TRIGGER AS $$
begin
if NEW.name is null then
NEW.name := NEW.uri;
end if;
if NEW.uri is null then
NEW.uri := NEW.name;
end if;
return NEW;
end;
$$ LANGUAGE plpgsql;
CREATE TRIGGER trig_set_repo_name BEFORE INSERT ON repo FOR EACH ROW
  EXECUTE PROCEDURE set_repo_name();

-- Add back the unused columns
ALTER TABLE repo
  ADD COLUMN IF NOT EXISTS pushed_at timestamp with time zone,
  ADD COLUMN IF NOT EXISTS indexed_revision text,
  ADD COLUMN IF NOT EXISTS freeze_indexed_revision boolean;

COMMIT;
