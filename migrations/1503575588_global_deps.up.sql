CREATE table "global_dep" (
	"language" text NOT NULL,
	"dep_data" jsonb NOT NULL,
	"repo_id" integer NOT NULL,
	"hints" jsonb
);
CREATE INDEX "global_dep_idxgin" ON "global_dep" USING gin ("dep_data" jsonb_path_ops);
CREATE INDEX "global_dep_repo_id" ON "global_dep" USING btree ("repo_id");
CREATE INDEX "global_dep_language" ON "global_dep" USING btree ("language");

CREATE table "global_dep_private" (
	"language" text NOT NULL,
	"dep_data" jsonb NOT NULL,
	"repo_id" integer NOT NULL,
	"hints" jsonb
);
CREATE INDEX "global_dep_private_idxgin" ON "global_dep_private" USING gin ("dep_data" jsonb_path_ops);
CREATE INDEX "global_dep_private_repo_id" ON "global_dep_private" USING btree ("repo_id");
CREATE INDEX "global_dep_private_language" ON "global_dep_private" USING btree ("language");

CREATE table "pkgs" (
	"repo_id" integer NOT NULL,
	"language" text NOT NULL,
	"pkg" jsonb NOT NULL
);
CREATE INDEX "pkg_pkg_idx" ON "pkgs" USING gin ("pkg" "jsonb_path_ops");
CREATE INDEX "pkg_lang_idx" ON "pkgs" USING btree ("language");
CREATE INDEX "pkg_repo_idx" ON "pkgs" USING btree ("repo_id");
