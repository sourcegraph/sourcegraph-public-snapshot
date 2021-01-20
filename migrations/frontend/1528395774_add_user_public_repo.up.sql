BEGIN;

CREATE TABLE IF NOT EXISTS user_public_repos (
    user_id integer,
    repo_id integer,
    UNIQUE (user_id, repo_id),
    CONSTRAINT user_fk
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE,
    CONSTRAINT repo_fk
        FOREIGN KEY (repo_id)
        REFERENCES repo (id)
        ON DELETE CASCADE
);

COMMIT;
