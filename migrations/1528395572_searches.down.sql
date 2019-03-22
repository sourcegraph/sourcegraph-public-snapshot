BEGIN;

DROP TRIGGER IF EXISTS trigger_delete_old_rows ON searches;
DROP FUNCTION IF EXISTS delete_old_rows();
DROP TABLE IF EXISTS searches;

COMMIT;
