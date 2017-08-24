CREATE TABLE repo (
	"id" SERIAL PRIMARY KEY,
	"uri" citext,
	"owner" citext,
	"name" citext,
	"description" text,
	"vcs" text NOT NULL,
	"http_clone_url" text,
	"ssh_clone_url" text,
	"homepage_url" text,
	"default_branch" text NOT NULL,
	"language" text,
	"blocked" boolean,
	"deprecated" boolean,
	"fork" boolean,
	"mirror" boolean,
	"private" boolean,
	"created_at" timestamp with time zone,
	"updated_at" timestamp with time zone,
	"pushed_at" timestamp with time zone,
	"vcs_synced_at" timestamp with time zone,
	"indexed_revision" text,
	"freeze_indexed_revision" boolean,
	"origin_repo_id" text,
	"origin_service" integer,
	"origin_api_base_url" text
);

CREATE UNIQUE INDEX "repo_uri_unique" ON repo(uri);
CREATE INDEX "repo_name" ON repo(name text_pattern_ops);
CREATE INDEX "repo_owner_ci" ON repo(owner);
CREATE INDEX "repo_name_ci" ON repo(name);
CREATE INDEX "repo_uri_trgm" ON repo USING GIN (lower(uri) gin_trgm_ops);
