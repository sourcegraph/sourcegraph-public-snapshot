CREATE TABLE IF NOT EXISTS repo_history(
    id BIGSERIAL NOT NULL PRIMARY KEY,
    repo_id INTEGER NOT NULL,
    timestamp DATE NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    message VARCHAR(255),
    metadata VARCHAR(255),
    CONSTRAINT fk_repo
        FOREIGN KEY(repo_id)
            REFERENCES repo(id)
);
