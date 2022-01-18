BEGIN;

DROP INDEX IF EXISTS notebook_stars_notebook_id_user_id_unique;

DROP INDEX IF EXISTS notebook_stars_notebook_id_idx;

DROP INDEX IF EXISTS notebook_stars_user_id_idx;

DROP TABLE IF EXISTS notebook_stars;

COMMIT;
