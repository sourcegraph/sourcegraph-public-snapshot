CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE "threads" (
	"id" bigserial NOT NULL PRIMARY KEY, 
	"local_repo_id" bigint, "file" text,
	"revision" text, "start_line" integer,
	"end_line" integer,
	"start_character" integer,
	"end_character" integer,
	"created_at" TIMESTAMP WITH TIME ZONE
);

CREATE TABLE "comments" (
	"id" bigserial NOT NULL PRIMARY KEY,
	"thread_id" bigint,
	"contents" text,
	"created_at" TIMESTAMP WITH TIME ZONE,
	"updated_at" TIMESTAMP WITH TIME ZONE,
	"deleted_at" TIMESTAMP WITH TIME ZONE,
	"author_name" text, "author_email" text
);

CREATE TABLE "local_repos" (
	"id" bigserial NOT NULL PRIMARY KEY,
	"remote_uri" citext,
	"access_token" text,
	"created_at" TIMESTAMP WITH TIME ZONE,
	"updated_at" TIMESTAMP WITH TIME ZONE,
	"deleted_at" TIMESTAMP WITH TIME ZONE,
	UNIQUE ("remote_uri", "access_token")
);

CREATE INDEX ON local_repos(remote_uri);
CREATE INDEX ON threads(local_repo_id, file);
