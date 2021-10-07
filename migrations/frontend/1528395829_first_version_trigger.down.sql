BEGIN;

DROP TRIGGER versions_insert ON versions;
DROP FUNCTION versions_insert_row_trigger;

COMMIT;
