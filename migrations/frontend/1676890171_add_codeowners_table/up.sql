CREATE TABLE IF NOT EXISTS codeowners (
	id SERIAL PRIMARY KEY,
    contents text,
    contents_proto json,
    repo_id int unique
   );