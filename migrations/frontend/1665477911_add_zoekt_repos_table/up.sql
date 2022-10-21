CREATE TABLE IF NOT EXISTS zoekt_repos (
    repo_id integer NOT NULL PRIMARY KEY REFERENCES repo(id) ON DELETE CASCADE,
    branches jsonb DEFAULT '[]'::jsonb NOT NULL,

    index_status text DEFAULT 'not_indexed'::text NOT NULL,

    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE INDEX IF NOT EXISTS zoekt_repos_index_status ON zoekt_repos USING btree (index_status);

CREATE OR REPLACE FUNCTION func_insert_zoekt_repo() RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO zoekt_repos (repo_id) VALUES (NEW.id);

  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trig_create_zoekt_repo_on_repo_insert on repo;

CREATE TRIGGER trig_create_zoekt_repo_on_repo_insert
AFTER INSERT
ON repo
FOR EACH ROW
EXECUTE FUNCTION func_insert_zoekt_repo();
