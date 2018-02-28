CREATE table "global_dep_private" (
  "language" text NOT NULL,
  "dep_data" jsonb NOT NULL,
  "repo_id" integer NOT NULL,
  "hints" jsonb
);
