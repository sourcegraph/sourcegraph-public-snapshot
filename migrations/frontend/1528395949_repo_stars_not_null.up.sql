BEGIN;

CREATE OR REPLACE FUNCTION set_repo_stars_null_to_zero()
RETURNS void AS
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

    SELECT COUNT(*) INTO remaining FROM repo WHERE stars IS NULL;

    RAISE NOTICE 'repo_stars_not_null.up.sql: % remaining', remaining;
  END LOOP;
END
$BODY$
LANGUAGE plpgsql;

SELECT set_repo_stars_null_to_zero();

ALTER TABLE repo
  ALTER COLUMN stars SET NOT NULL,
  ALTER COLUMN stars SET DEFAULT 0;
