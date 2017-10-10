CREATE TABLE "org_settings" (
	"id" serial NOT NULL PRIMARY KEY,
	"org_id" integer NOT NULL,
	"author_auth0_id" text NOT NULL, 
	"contents" text,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),

	CONSTRAINT org_settings_references_orgs FOREIGN KEY (org_id) REFERENCES orgs (id) ON DELETE RESTRICT,
	CONSTRAINT org_settings_references_users FOREIGN KEY (author_auth0_id) REFERENCES users (auth0_id) ON DELETE RESTRICT
);