BEGIN;
ALTER TABLE repo RENAME COLUMN blocked TO enabled;
-- Renaming from blocked to enabled inverts the meaning.
UPDATE repo SET enabled=(enabled=false OR enabled IS NULL);
ALTER TABLE repo ALTER COLUMN enabled SET NOT NULL;
ALTER TABLE repo ALTER COLUMN enabled SET DEFAULT true;
COMMIT;

