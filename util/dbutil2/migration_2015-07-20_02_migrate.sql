-- WARNING: Do not run this unless you know what you are doing, this is
-- recorded for historical reasons. Please see the 2015-07-20 section of
-- MIGRATION.md

-- repo
INSERT INTO repo
        (uri, origin, name, description, vcs, http_clone_url, ssh_clone_url, homepage_url, default_branch, language, blocked, deprecated, fork, mirror, private, created_at, updated_at, pushed_at)
SELECT
        uri, -- uri
        substring(uri from '[^/]+'), -- origin, extract first part of uri. eg github.com from github.com/gorilla/mux
        name, -- name
        description, -- description
        vcs, -- vcs
        http_clone_url, -- http_clone_url
        ssh_clone_url, -- ssh_clone_url
        homepage_url, -- homepage_url
        default_branch, -- default_branch
        language, -- language
        FALSE, -- blocked
        deprecated, -- deprecated
        fork, -- fork
        mirror, -- mirror
        private, -- private
        created_at, -- created_at
        updated_at, -- updated_at
        pushed_at -- pushed_at
FROM old_repo;

UPDATE repo SET mirror=false, origin=NULL WHERE origin='sourcegraph.com';
UPDATE repo SET mirror=true WHERE origin IS NOT NULL;


-- repo_build
INSERT INTO repo_build
        (attempt, repo, commit_id, created_at, started_at, ended_at, heartbeat_at, success, failure, killed, host, purged, import, queue, usecache, priority)
SELECT
        1, -- attempt
        (SELECT uri FROM old_repo WHERE rid = old.repo), -- repo
        commit_id, -- commit_id
        created_at, -- created_at
        started_at, -- started_at
        ended_at, -- ended_at
        heartbeat_at, -- heartbeat_at
        success, -- success
        failure, -- failure
        killed, -- killed
        host, -- host
        purged, -- purged
        import, -- import
        queue, -- queue
        use_cache, -- usecache
        priority -- priority
FROM old_repo_build as old
WHERE started_at = (SELECT max(started_at) FROM old_repo_build WHERE repo = old.repo AND commit_id = old.commit_id)
      AND (SELECT uri FROM old_repo WHERE rid = old.repo) IS NOT NULL;

-- repo_build_task
INSERT INTO repo_build_task
        (taskid, repo, attempt, commit_id, unit_type, unit, op, "order", created_at, started_at, ended_at, queue, success, failure)
SELECT
        task.taskid, -- taskid
        new_build.repo, -- repo
        new_build.attempt, -- attempt
        new_build.commit_id, -- commit_id
        task.unit_type, -- unit_type
        task.unit, -- unit
        task.op, -- op
        task."order", -- order
        task.created_at, -- created_at
        task.started_at, -- started_at
        task.ended_at, -- ended_at
        task.queue, -- queue
        task.success, -- success
        task.failure -- failure
FROM old_repo_build_task as task, old_repo_build as old_build, repo_build as new_build, old_repo as repo
WHERE task.bid = old_build.bid AND new_build.attempt = 1 AND new_build.commit_id = old_build.commit_id AND old_build.repo = repo.rid AND new_build.repo = repo.uri;

-- repo_hit
INSERT INTO repo_hit
        (repo, at)
SELECT
        (SELECT uri FROM old_repo WHERE rid = old.repo), -- repo
        at -- at
FROM old_repo_hit AS old;

-- repo_key
INSERT INTO repo_key
        (repo, private_key_pem)
SELECT
        repo.uri, -- repo
        old.private_key_pem -- private_key_pem
FROM old_repo_key AS old, old_repo AS repo
WHERE old.rid = repo.rid;
