CREATE OR REPLACE PROCEDURE set_repo_stars_null_to_zero() AS
$BODY$
DECLARE
  remaining integer;

BEGIN
  SELECT COUNT(*) INTO remaining FROM repo WHERE stars IS NULL;
  WHILE remaining > 0 LOOP
    UPDATE repo SET stars = 0
    FROM (
      SELECT id FROM repo
      WHERE stars IS NULL
      LIMIT 10000
      FOR UPDATE SKIP LOCKED
    ) s
    WHERE repo.id = s.id;

    COMMIT;

    SELECT COUNT(*) INTO remaining FROM repo WHERE stars IS NULL;

    RAISE NOTICE 'repo_stars_not_null.up.sql: % remaining', remaining;
  END LOOP;
END
$BODY$
LANGUAGE plpgsql;
