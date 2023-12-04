CREATE TABLE IF NOT EXISTS codeowners (
    id SERIAL PRIMARY KEY,
    contents text not null,
    contents_proto bytea not null,
    repo_id int unique not null,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),

    CONSTRAINT codeowners_repo_id_fkey
        FOREIGN KEY (repo_id) REFERENCES repo (id) ON DELETE CASCADE
);