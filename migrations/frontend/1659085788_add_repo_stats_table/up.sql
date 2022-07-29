CREATE TABLE repo_statistics (
  -- We only allow one row in this table.
  id bool PRIMARY KEY DEFAULT TRUE,
  -- Constraint ensures that the `id` must be true and `PRIMARY KEY` ensures
  -- that it's unique.
  CONSTRAINT id CHECK (id),

  total bigint,
  soft_deleted bigint,
  cloned bigint
);

COMMENT ON COLUMN repo_statistics.total IS 'Number of repositories that are not soft-deleted and not blocked';
COMMENT ON COLUMN repo_statistics.soft_deleted IS 'Number of repositories that are soft-deleted and not blocked';
COMMENT ON COLUMN repo_statistics.cloned IS 'Number of repositories that are cloned';

INSERT INTO repo_statistics (total, soft_deleted, cloned) VALUES (0, 0, 0);

UPDATE repo_statistics
SET
  total = (SELECT COUNT(1) FROM repo WHERE deleted_at is NULL AND blocked IS NULL),
  soft_deleted = (SELECT COUNT(1) FROM repo WHERE deleted_at is NOT NULL AND blocked IS NULL),
  cloned = (SELECT COUNT(1) FROM gitserver_repos WHERE clone_status = 'cloned')
WHERE
  id = TRUE;

CREATE FUNCTION count_soft_deleted_repo() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
    BEGIN
        -- if a repo is soft-deleted, delete every row that references that repo
        IF (OLD.deleted_at IS NULL AND OLD.blocked IS NULL AND NEW.deleted_at IS NOT NULL AND NEW.blocked IS NULL) THEN
          UPDATE repo_statistics
          SET soft_deleted = soft_deleted + 1, total = total - 1
          WHERE id = TRUE;
          END IF;
        RETURN OLD;
    END;
$$;
CREATE TRIGGER trig_count_soft_deleted_repo AFTER UPDATE OF deleted_at ON repo FOR EACH ROW EXECUTE FUNCTION count_soft_deleted_repo();

CREATE FUNCTION count_inserted_repo() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
    BEGIN
        IF (NEW.deleted_at IS NOT NULL AND NEW.blocked IS NULL) THEN
          UPDATE repo_statistics
          SET soft_deleted = soft_deleted + 1
          WHERE id = TRUE;
        ELSE
          UPDATE repo_statistics
          SET total = total + 1
          WHERE id = TRUE;
        END IF;
        RETURN NEW;
    END;
$$;
CREATE TRIGGER trig_count_inserted_repo AFTER INSERT ON repo FOR EACH ROW EXECUTE FUNCTION count_inserted_repo();

CREATE FUNCTION count_deleted_repo() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
    BEGIN
        IF (NEW.blocked IS NULL) THEN
          IF (NEW.deleted_at IS NOT NULL) THEN
            UPDATE repo_statistics
            SET soft_deleted = soft_deleted - 1
            WHERE id = TRUE;
          ELSIF (NEW.deleted_at IS NULL) THEN
            UPDATE repo_statistics
            SET total = total - 1
            WHERE id = TRUE;
          END IF;
        END IF;

        RETURN NULL;
    END;
$$;
CREATE TRIGGER trig_count_deleted_repo AFTER DELETE ON repo FOR EACH ROW EXECUTE FUNCTION count_deleted_repo();

CREATE FUNCTION count_cloned_gitserver_repos() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
    BEGIN
        -- if a repo is soft-deleted, delete every row that references that repo
        IF ((OLD.clone_status = 'cloning' OR OLD.clone_status = '') AND NEW.clone_status = 'cloned') THEN
          UPDATE repo_statistics
          SET
            cloned = cloned + 1
          WHERE
              id = TRUE;
        ELSIF (OLD.clone_status = 'cloned' AND NEW.clone_status = 'not_cloned') THEN
          UPDATE repo_statistics
          SET
            cloned = cloned - 1
          WHERE
              id = TRUE;
          END IF;
        RETURN OLD;
    END;
$$;
CREATE TRIGGER trig_count_cloned_gitserver_repos AFTER UPDATE OF clone_status ON gitserver_repos FOR EACH ROW EXECUTE FUNCTION count_cloned_gitserver_repos();

CREATE FUNCTION count_deleted_gitserver_repos() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
    BEGIN
        IF (NEW.clone_status = 'cloned') THEN
          UPDATE repo_statistics
          SET
            cloned = cloned - 1
          WHERE
              id = TRUE;
        END IF;

        RETURN NULL;
    END;
$$;
CREATE TRIGGER trig_count_deleted_gitserver_repos AFTER DELETE ON gitserver_repos FOR EACH ROW EXECUTE FUNCTION count_deleted_gitserver_repos();
