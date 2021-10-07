BEGIN;

CREATE OR REPLACE FUNCTION versions_insert_row_trigger() RETURNS trigger LANGUAGE plpgsql AS $lang$
BEGIN
    NEW.first_version = NEW.version;
    RETURN NEW;
END $lang$;

CREATE TRIGGER versions_insert BEFORE INSERT ON versions FOR EACH ROW EXECUTE PROCEDURE versions_insert_row_trigger();

COMMIT;
