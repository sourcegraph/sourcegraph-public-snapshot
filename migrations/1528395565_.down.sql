CREATE table "global_dep" (
	"language" text NOT NULL,
	"dep_data" jsonb NOT NULL,
	"repo_id" integer NOT NULL,
	"hints" jsonb
);
CREATE INDEX "global_dep_idxgin" ON "global_dep" USING gin ("dep_data" jsonb_path_ops);
CREATE INDEX "global_dep_repo_id" ON "global_dep" USING btree ("repo_id");
CREATE INDEX "global_dep_language" ON "global_dep" USING btree ("language");
CREATE INDEX "global_dep_idx_package" ON "global_dep" ((dep_data ->> 'package' COLLATE "C"));

ALTER TABLE global_dep ADD CONSTRAINT global_dep_repo_id FOREIGN KEY (repo_id) REFERENCES repo (id) ON DELETE RESTRICT;

CREATE table "pkgs" (
	"repo_id" integer NOT NULL,
	"language" text NOT NULL,
	"pkg" jsonb NOT NULL
);
CREATE INDEX "pkg_pkg_idx" ON "pkgs" USING gin ("pkg" "jsonb_path_ops");
CREATE INDEX "pkg_lang_idx" ON "pkgs" USING btree ("language");
CREATE INDEX "pkg_repo_idx" ON "pkgs" USING btree ("repo_id");

ALTER TABLE pkgs ADD CONSTRAINT pkgs_repo_id FOREIGN KEY (repo_id) REFERENCES repo (id) ON DELETE RESTRICT;
