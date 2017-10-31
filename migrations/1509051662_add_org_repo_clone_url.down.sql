BEGIN;

ALTER TABLE "org_repos"
	RENAME "canonical_remote_id" TO "remote_uri";
ALTER TABLE "org_repos"
	DROP COLUMN "clone_url";

COMMIT;
