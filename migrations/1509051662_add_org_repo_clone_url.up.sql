BEGIN;

ALTER TABLE "org_repos" ADD COLUMN "clone_url" text;

-- Backfill clone_url column with former remote_uri data. 
-- Prepend the correct vendor-specific schemes (github:// for github.com, 
-- bitbucketcloud:// for bitbucket.org) and assume the rest are https.
UPDATE "org_repos"
	SET clone_url = (CASE
		WHEN remote_uri LIKE 'github.com%' THEN concat('github://', remote_uri)
		WHEN remote_uri LIKE 'bitbucket.org%' THEN concat('bitbucketcloud://', remote_uri)
		ELSE concat('https://', remote_uri)
	END);

ALTER TABLE "org_repos"
	ALTER COLUMN "clone_url" SET NOT NULL,
	ADD CONSTRAINT clone_url_valid CHECK (clone_url ~ '^([^\s]+://)?[^\s]+$');

ALTER TABLE "org_repos"
	RENAME "remote_uri" TO "canonical_remote_id";

COMMIT;
