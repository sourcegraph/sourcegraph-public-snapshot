CREATE TABLE repo_statistics (
  -- We only allow one row in this table.
  id bool PRIMARY KEY DEFAULT TRUE,
  -- Constraint ensures that the `id` must be true and `PRIMARY KEY` ensures
  -- that it's unique.
  CONSTRAINT id CHECK (id),

  total BIGINT NOT NULL DEFAULT 0,
  soft_deleted BIGINT NOT NULL DEFAULT 0,
  cloned BIGINT NOT NULL DEFAULT 0
);

COMMENT ON COLUMN repo_statistics.total IS 'Number of repositories that are not soft-deleted and not blocked';
COMMENT ON COLUMN repo_statistics.soft_deleted IS 'Number of repositories that are soft-deleted and not blocked';
COMMENT ON COLUMN repo_statistics.cloned IS 'Number of repositories that are cloned';

INSERT INTO repo_statistics (total, soft_deleted, cloned)
VALUES (
  (SELECT COUNT(1) FROM repo WHERE deleted_at is NULL AND blocked IS NULL),
  (SELECT COUNT(1) FROM repo WHERE deleted_at is NOT NULL AND blocked IS NULL),
  (SELECT COUNT(1) FROM gitserver_repos WHERE clone_status = 'cloned')
);

--------------------------------------------------------------------------------
--                              repos table                                   --
--------------------------------------------------------------------------------
-- UPDATE
CREATE OR REPLACE FUNCTION recalc_repo_statistics_on_repo_update() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
      INSERT INTO
        repo_statistics (total, soft_deleted)
      VALUES (
        (SELECT COUNT(1) FROM newtab WHERE deleted_at IS NULL     AND blocked IS NULL) - (SELECT COUNT(1) FROM oldtab WHERE deleted_at IS NULL     AND blocked IS NULL),
        (SELECT COUNT(1) FROM newtab WHERE deleted_at IS NOT NULL AND blocked IS NULL) - (SELECT COUNT(1) FROM oldtab WHERE deleted_at IS NOT NULL AND blocked IS NULL)
      )
      ON CONFLICT(id) DO UPDATE
      SET
        total        = repo_statistics.total        + excluded.total,
        soft_deleted = repo_statistics.soft_deleted + excluded.soft_deleted
      ;
      RETURN NULL;
  END
$$;
DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_repo_update ON repo;
CREATE TRIGGER trig_recalc_repo_statistics_on_repo_update AFTER UPDATE ON repo REFERENCING OLD TABLE AS oldtab NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION recalc_repo_statistics_on_repo_update();

-- INSERT
CREATE OR REPLACE FUNCTION recalc_repo_statistics_on_repo_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
      INSERT INTO
        repo_statistics (total, soft_deleted)
      VALUES (
        (SELECT COUNT(1) FROM newtab WHERE deleted_at IS NULL     AND blocked IS NULL),
        (SELECT COUNT(1) FROM newtab WHERE deleted_at IS NOT NULL AND blocked IS NULL)
      )
      ON CONFLICT(id) DO UPDATE
      SET
        total        = repo_statistics.total        + excluded.total,
        soft_deleted = repo_statistics.soft_deleted + excluded.soft_deleted
      ;
      RETURN NULL;
  END
$$;
DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_repo_insert ON repo;
CREATE TRIGGER trig_recalc_repo_statistics_on_repo_insert AFTER INSERT ON repo REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION recalc_repo_statistics_on_repo_insert();

-- DELETE
CREATE OR REPLACE FUNCTION recalc_repo_statistics_on_repo_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
      UPDATE repo_statistics
      SET
        total        = repo_statistics.total        - (SELECT COUNT(1) FROM oldtab WHERE deleted_at IS NULL     AND blocked IS NULL),
        soft_deleted = repo_statistics.soft_deleted - (SELECT COUNT(1) FROM oldtab WHERE deleted_at IS NOT NULL AND blocked IS NULL)
      WHERE id = TRUE;
      RETURN NULL;
  END
$$;
DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_repo_delete ON repo;
CREATE TRIGGER trig_recalc_repo_statistics_on_repo_delete AFTER DELETE ON repo REFERENCING OLD TABLE AS oldtab FOR EACH STATEMENT EXECUTE FUNCTION recalc_repo_statistics_on_repo_delete();



--------------------------------------------------------------------------------
--                       gitserver_repos table                                --
--------------------------------------------------------------------------------
-- UPDATE
CREATE OR REPLACE FUNCTION recalc_repo_statistics_on_gitserver_repos_update() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
      INSERT INTO
        repo_statistics (cloned)
      VALUES (
        (SELECT COUNT(1) FROM newtab WHERE clone_status = 'cloned') - (SELECT COUNT(1) FROM oldtab WHERE clone_status = 'cloned')
      )
      ON CONFLICT(id) DO UPDATE
      SET
        cloned = repo_statistics.cloned + excluded.cloned
      ;
      RETURN NULL;
  END
$$;

DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_gitserver_repos_update ON gitserver_repos;
CREATE TRIGGER trig_recalc_repo_statistics_on_gitserver_repos_update AFTER UPDATE ON gitserver_repos REFERENCING OLD TABLE AS oldtab NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION recalc_repo_statistics_on_gitserver_repos_update();


-- INSERT
CREATE OR REPLACE FUNCTION recalc_repo_statistics_on_gitserver_repos_insert() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
      INSERT INTO
        repo_statistics (cloned)
      VALUES (
        (SELECT COUNT(1) FROM newtab WHERE clone_status = 'cloned')
      )
      ON CONFLICT(id) DO UPDATE
      SET
        cloned = repo_statistics.cloned + excluded.cloned
      ;
      RETURN NULL;
  END
$$;
DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_gitserver_repos_insert ON gitserver_repos;
CREATE TRIGGER trig_recalc_repo_statistics_on_gitserver_repos_insert AFTER INSERT ON gitserver_repos REFERENCING NEW TABLE AS newtab FOR EACH STATEMENT EXECUTE FUNCTION recalc_repo_statistics_on_gitserver_repos_insert();

-- DELETE
CREATE OR REPLACE FUNCTION recalc_repo_statistics_on_gitserver_repos_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$ BEGIN
      UPDATE repo_statistics
      SET
        cloned = repo_statistics.cloned - (SELECT COUNT(1) FROM oldtab WHERE clone_status = 'cloned')
      WHERE id = TRUE;
      RETURN NULL;
  END
$$;
DROP TRIGGER IF EXISTS trig_recalc_repo_statistics_on_gitserver_repos_delete ON gitserver_repos;
CREATE TRIGGER trig_recalc_repo_statistics_on_gitserver_repos_delete AFTER DELETE ON gitserver_repos REFERENCING OLD TABLE AS oldtab FOR EACH STATEMENT EXECUTE FUNCTION recalc_repo_statistics_on_gitserver_repos_delete();
