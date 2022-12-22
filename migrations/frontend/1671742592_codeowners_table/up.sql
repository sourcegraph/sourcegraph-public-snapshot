CREATE TABLE IF NOT EXISTS codeowners_head (
  repo_id integer PRIMARY KEY REFERENCES repo(id) ON DELETE CASCADE DEFERRABLE,
  proto bytea NOT NULL,
  updated_at timestamp with time zone DEFAULT now() NOT NULL,
  created_at timestamp with time zone DEFAULT now() NOT NULL
);
