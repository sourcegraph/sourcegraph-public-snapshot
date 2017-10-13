BEGIN;

ALTER TABLE "threads" RENAME COLUMN "repo_revision" to "revision";

ALTER TABLE "threads"
	ALTER COLUMN "revision" DROP NOT NULL,
	DROP COLUMN "lines_revision";

COMMIT;
