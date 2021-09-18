BEGIN;

CREATE OR REPLACE FUNCTION versions_insert_row_trigger() RETURNS trigger LANGUAGE plpgsql AS $lang$
BEGIN
    NEW.first_version = NEW.version;
    RETURN NEW;
END $lang$;

CREATE TRIGGER versions_insert BEFORE INSERT ON versions FOR EACH ROW EXECUTE PROCEDURE versions_insert_row_trigger();

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
