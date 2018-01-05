BEGIN;
ALTER TABLE repo RENAME COLUMN enabled TO blocked;
ALTER TABLE repo ALTER COLUMN blocked DROP NOT NULL;
-- Invert the meaning.
UPDATE repo SET blocked=null WHERE blocked;
UPDATE repo SET blocked=true WHERE blocked=false;
COMMIT;
