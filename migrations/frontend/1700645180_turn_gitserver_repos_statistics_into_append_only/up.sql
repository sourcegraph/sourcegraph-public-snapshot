-- Remove primary key so we can have more rows per shard_id
ALTER TABLE gitserver_repos_statistics
DROP CONSTRAINT gitserver_repos_statistics_pkey;

-- Add index
CREATE INDEX gitserver_repos_statistics_shard_id ON gitserver_repos_statistics(shard_id);

------------------------------------------------------------------
-- INSERT
------------------------------------------------------------------
CREATE OR REPLACE FUNCTION recalc_gitserver_repos_statistics_on_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
      -------------------------------------------------
      -- THIS IS CHANGED TO APPEND
      -------------------------------------------------
      INSERT INTO gitserver_repos_statistics AS grs (shard_id, total, not_cloned, cloning, cloned, failed_fetch)
      SELECT
        shard_id,
        COUNT(*) AS total,
        COUNT(*) FILTER(WHERE clone_status = 'not_cloned') AS not_cloned,
        COUNT(*) FILTER(WHERE clone_status = 'cloning') AS cloning,
        COUNT(*) FILTER(WHERE clone_status = 'cloned') AS cloned,
        COUNT(*) FILTER(WHERE last_error IS NOT NULL) AS failed_fetch
      FROM
        newtab
      GROUP BY shard_id
      ;

      RETURN NULL;
  END
$$;

------------------------------------------------------------------
-- UPDATE
------------------------------------------------------------------
CREATE OR REPLACE FUNCTION recalc_gitserver_repos_statistics_on_update() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN

      -------------------------------------------------
      -- THIS IS CHANGED TO APPEND
      -------------------------------------------------
      WITH diff(shard_id, total, not_cloned, cloning, cloned, failed_fetch, corrupted) AS (
        SELECT
            COALESCE(newtab.shard_id, oldtab.shard_id) AS shard_id,
            COUNT(newtab.repo_id) - COUNT(oldtab.repo_id) AS total,
            COUNT(newtab.repo_id) FILTER (WHERE newtab.clone_status = 'not_cloned') - COUNT(oldtab.repo_id) FILTER (WHERE oldtab.clone_status = 'not_cloned') AS not_cloned,
            COUNT(newtab.repo_id) FILTER (WHERE newtab.clone_status = 'cloning')    - COUNT(oldtab.repo_id) FILTER (WHERE oldtab.clone_status = 'cloning') AS cloning,
            COUNT(newtab.repo_id) FILTER (WHERE newtab.clone_status = 'cloned')     - COUNT(oldtab.repo_id) FILTER (WHERE oldtab.clone_status = 'cloned') AS cloned,
            COUNT(newtab.repo_id) FILTER (WHERE newtab.last_error IS NOT NULL)      - COUNT(oldtab.repo_id) FILTER (WHERE oldtab.last_error IS NOT NULL) AS failed_fetch,
            COUNT(newtab.repo_id) FILTER (WHERE newtab.corrupted_at IS NOT NULL)    - COUNT(oldtab.repo_id) FILTER (WHERE oldtab.corrupted_at IS NOT NULL) AS corrupted
        FROM
            newtab
        FULL OUTER JOIN
            oldtab ON newtab.repo_id = oldtab.repo_id AND newtab.shard_id = oldtab.shard_id
        GROUP BY
            COALESCE(newtab.shard_id, oldtab.shard_id)
      )
      INSERT INTO gitserver_repos_statistics AS grs (shard_id, total, not_cloned, cloning, cloned, failed_fetch, corrupted)
      SELECT shard_id, total, not_cloned, cloning, cloned, failed_fetch, corrupted
      FROM diff
      WHERE
            total != 0
        OR not_cloned != 0
        OR cloning != 0
        OR cloned != 0
        OR failed_fetch != 0
        OR corrupted != 0
      ;

      -------------------------------------------------
      -- UNCHANGED
      -------------------------------------------------
      WITH diff(not_cloned, cloning, cloned, failed_fetch, corrupted) AS (
        VALUES (
          (
            (SELECT COUNT(*) FROM newtab JOIN repo r ON newtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND newtab.clone_status = 'not_cloned')
            -
            (SELECT COUNT(*) FROM oldtab JOIN repo r ON oldtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND oldtab.clone_status = 'not_cloned')
          ),
          (
            (SELECT COUNT(*) FROM newtab JOIN repo r ON newtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND newtab.clone_status = 'cloning')
            -
            (SELECT COUNT(*) FROM oldtab JOIN repo r ON oldtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND oldtab.clone_status = 'cloning')
          ),
          (
            (SELECT COUNT(*) FROM newtab JOIN repo r ON newtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND newtab.clone_status = 'cloned')
            -
            (SELECT COUNT(*) FROM oldtab JOIN repo r ON oldtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND oldtab.clone_status = 'cloned')
          ),
          (
            (SELECT COUNT(*) FROM newtab JOIN repo r ON newtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND newtab.last_error IS NOT NULL)
            -
            (SELECT COUNT(*) FROM oldtab JOIN repo r ON oldtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND oldtab.last_error IS NOT NULL)
          ),
          (
            (SELECT COUNT(*) FROM newtab JOIN repo r ON newtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND newtab.corrupted_at IS NOT NULL)
            -
            (SELECT COUNT(*) FROM oldtab JOIN repo r ON oldtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND oldtab.corrupted_at IS NOT NULL)
          )

        )
      )
      INSERT INTO repo_statistics (not_cloned, cloning, cloned, failed_fetch, corrupted)
      SELECT not_cloned, cloning, cloned, failed_fetch, corrupted
      FROM diff
      WHERE
           not_cloned != 0
        OR cloning != 0
        OR cloned != 0
        OR failed_fetch != 0
        OR corrupted != 0
      ;

      RETURN NULL;
  END
$$;


------------------------------------------------------------------
-- DELETE
------------------------------------------------------------------
CREATE OR REPLACE FUNCTION recalc_gitserver_repos_statistics_on_update() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN

      -------------------------------------------------
      -- THIS IS CHANGED TO APPEND
      -------------------------------------------------
      WITH diff(shard_id, total, not_cloned, cloning, cloned, failed_fetch, corrupted) AS (
        SELECT
            COALESCE(newtab.shard_id, oldtab.shard_id) AS shard_id,
            COUNT(newtab.repo_id) - COUNT(oldtab.repo_id) AS total,
            COUNT(newtab.repo_id) FILTER (WHERE newtab.clone_status = 'not_cloned') - COUNT(oldtab.repo_id) FILTER (WHERE oldtab.clone_status = 'not_cloned') AS not_cloned,
            COUNT(newtab.repo_id) FILTER (WHERE newtab.clone_status = 'cloning')    - COUNT(oldtab.repo_id) FILTER (WHERE oldtab.clone_status = 'cloning') AS cloning,
            COUNT(newtab.repo_id) FILTER (WHERE newtab.clone_status = 'cloned')     - COUNT(oldtab.repo_id) FILTER (WHERE oldtab.clone_status = 'cloned') AS cloned,
            COUNT(newtab.repo_id) FILTER (WHERE newtab.last_error IS NOT NULL)      - COUNT(oldtab.repo_id) FILTER (WHERE oldtab.last_error IS NOT NULL) AS failed_fetch,
            COUNT(newtab.repo_id) FILTER (WHERE newtab.corrupted_at IS NOT NULL)    - COUNT(oldtab.repo_id) FILTER (WHERE oldtab.corrupted_at IS NOT NULL) AS corrupted
        FROM
            newtab
        FULL OUTER JOIN
            oldtab ON newtab.repo_id = oldtab.repo_id AND newtab.shard_id = oldtab.shard_id
        GROUP BY
            COALESCE(newtab.shard_id, oldtab.shard_id)
      )
      INSERT INTO gitserver_repos_statistics AS grs (shard_id, total, not_cloned, cloning, cloned, failed_fetch, corrupted)
      SELECT shard_id, total, not_cloned, cloning, cloned, failed_fetch, corrupted
      FROM diff
      WHERE
            total != 0
        OR not_cloned != 0
        OR cloning != 0
        OR cloned != 0
        OR failed_fetch != 0
        OR corrupted != 0
      ;

      -------------------------------------------------
      -- UNCHANGED
      -------------------------------------------------
      WITH diff(not_cloned, cloning, cloned, failed_fetch, corrupted) AS (
        VALUES (
          (
            (SELECT COUNT(*) FROM newtab JOIN repo r ON newtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND newtab.clone_status = 'not_cloned')
            -
            (SELECT COUNT(*) FROM oldtab JOIN repo r ON oldtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND oldtab.clone_status = 'not_cloned')
          ),
          (
            (SELECT COUNT(*) FROM newtab JOIN repo r ON newtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND newtab.clone_status = 'cloning')
            -
            (SELECT COUNT(*) FROM oldtab JOIN repo r ON oldtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND oldtab.clone_status = 'cloning')
          ),
          (
            (SELECT COUNT(*) FROM newtab JOIN repo r ON newtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND newtab.clone_status = 'cloned')
            -
            (SELECT COUNT(*) FROM oldtab JOIN repo r ON oldtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND oldtab.clone_status = 'cloned')
          ),
          (
            (SELECT COUNT(*) FROM newtab JOIN repo r ON newtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND newtab.last_error IS NOT NULL)
            -
            (SELECT COUNT(*) FROM oldtab JOIN repo r ON oldtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND oldtab.last_error IS NOT NULL)
          ),
          (
            (SELECT COUNT(*) FROM newtab JOIN repo r ON newtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND newtab.corrupted_at IS NOT NULL)
            -
            (SELECT COUNT(*) FROM oldtab JOIN repo r ON oldtab.repo_id = r.id WHERE r.deleted_at is NULL AND r.blocked IS NULL AND oldtab.corrupted_at IS NOT NULL)
          )

        )
      )
      INSERT INTO repo_statistics (not_cloned, cloning, cloned, failed_fetch, corrupted)
      SELECT not_cloned, cloning, cloned, failed_fetch, corrupted
      FROM diff
      WHERE
           not_cloned != 0
        OR cloning != 0
        OR cloned != 0
        OR failed_fetch != 0
        OR corrupted != 0
      ;

      RETURN NULL;
  END
$$;

CREATE OR REPLACE FUNCTION recalc_gitserver_repos_statistics_on_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
      INSERT INTO gitserver_repos_statistics AS grs (shard_id, total, not_cloned, cloning, cloned, failed_fetch, corrupted)
      SELECT
        (SELECT oldtab.shard_id),
        (SELECT -COUNT(*)                                           FROM oldtab),
        (SELECT -COUNT(*) FILTER(WHERE clone_status = 'not_cloned') FROM oldtab),
        (SELECT -COUNT(*) FILTER(WHERE clone_status = 'cloning')    FROM oldtab),
        (SELECT -COUNT(*) FILTER(WHERE clone_status = 'cloned')     FROM oldtab),
        (SELECT -COUNT(*) FILTER(WHERE last_error IS NOT NULL)      FROM oldtab),
        (SELECT -COUNT(*) FILTER(WHERE corrupted_at IS NOT NULL)    FROM oldtab)
      FROM oldtab;

      RETURN NULL;
  END
$$;
