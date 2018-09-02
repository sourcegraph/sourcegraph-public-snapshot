-- Recreate user_tags and org_tags tables. These intentionally are not populated with any data
-- because these tables have not been used for 2+ major releases (instead, the users.tags column is
-- used).

CREATE TABLE "user_tags" (
	"id" serial PRIMARY KEY,
	"user_id" integer NOT NULL,
	"name" text NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE DEFAULT now(),
	"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT now(),
	"deleted_at" TIMESTAMP WITH TIME ZONE,
	CONSTRAINT user_tags_references_users FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT
);

CREATE TABLE "org_tags" (
	"id" serial PRIMARY KEY,
	"org_id" integer NOT NULL,
	"name" text NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE DEFAULT now(),
	"updated_at" TIMESTAMP WITH TIME ZONE DEFAULT now(),
	"deleted_at" TIMESTAMP WITH TIME ZONE,
	CONSTRAINT org_tags_references_users FOREIGN KEY (org_id) REFERENCES orgs (id) ON DELETE RESTRICT
);
