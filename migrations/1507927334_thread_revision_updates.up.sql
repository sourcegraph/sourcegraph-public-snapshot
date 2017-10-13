BEGIN;

ALTER TABLE "threads"
	ALTER COLUMN "revision" SET NOT NULL,
	ADD COLUMN "lines_revision" text;
ALTER TABLE "threads" RENAME COLUMN "revision" to "repo_revision";

UPDATE "threads" SET lines_revision=repo_revision;

ALTER TABLE "threads" ALTER COLUMN "lines_revision" SET NOT NULL;

COMMIT;
