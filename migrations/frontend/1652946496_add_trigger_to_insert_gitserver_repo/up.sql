-- batch insert repos which don't have a corresponding row in gitserver_repos table
INSERT INTO gitserver_repos(repo_id, shard_id)
SELECT repo.id, ''
FROM repo
LEFT JOIN gitserver_repos gr
ON repo.id = gr.repo_id
WHERE gr.repo_id IS NULL
ON CONFLICT (repo_id) DO NOTHING;

CREATE OR REPLACE FUNCTION func_insert_gitserver_repo() RETURNS TRIGGER AS $$
BEGIN
INSERT INTO gitserver_repos
(repo_id, shard_id)
VALUES (NEW.id, '');
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_gitserver_repo_insert on repo;

CREATE TRIGGER trigger_gitserver_repo_insert
    AFTER INSERT
    ON repo
    FOR EACH ROW
    EXECUTE FUNCTION func_insert_gitserver_repo();
