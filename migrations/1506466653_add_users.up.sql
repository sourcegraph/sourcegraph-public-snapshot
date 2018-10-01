CREATE TABLE "users" (
	"id" serial NOT NULL PRIMARY KEY,
	"auth0_id" text NOT NULL UNIQUE,
	"email" citext NOT NULL UNIQUE,
	"username" citext NOT NULL CONSTRAINT users_username_valid CHECK (username ~ '^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,37}[a-zA-Z0-9])?$') UNIQUE,
	"display_name" text NOT NULL,
	"avatar_url" text,
	"created_at" timestamp with time zone DEFAULT now(),
	"updated_at" timestamp with time zone DEFAULT now(),
	"deleted_at" timestamp
);
