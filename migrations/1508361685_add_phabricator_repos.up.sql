CREATE TABLE "phabricator_repos" (
	"id" serial PRIMARY KEY,
	"callsign" citext NOT NULL UNIQUE,
	"uri" citext NOT NULL UNIQUE,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	"updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	"deleted_at" TIMESTAMP WITH TIME ZONE
);
