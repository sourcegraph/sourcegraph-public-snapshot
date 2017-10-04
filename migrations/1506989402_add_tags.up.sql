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

ALTER TABLE orgs ADD COLUMN display_name text CONSTRAINT org_display_name_valid CHECK (char_length(display_name) <= 64);
