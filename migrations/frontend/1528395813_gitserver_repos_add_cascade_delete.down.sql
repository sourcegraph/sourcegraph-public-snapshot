BEGIN;

SET CONSTRAINTS ALL DEFERRED;

ALTER TABLE gitserver_repos
    DROP CONSTRAINT gitserver_repos_repo_id_fkey,
    ADD CONSTRAINT gitserver_repos_repo_id_fkey
        FOREIGN KEY (repo_id)
            REFERENCES repo (id);

COMMIT;
