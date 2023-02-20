CREATE TABLE IF NOT EXISTS codeowners (
	id SERIAL PRIMARY KEY,
    contents text,
    contents_proto json,
    repo_id int unique,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);