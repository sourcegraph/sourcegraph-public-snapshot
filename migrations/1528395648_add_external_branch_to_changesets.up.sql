BEGIN;

ALTER TABLE changesets ADD COLUMN external_branch TEXT;

UPDATE changesets SET external_branch = metadata -> 'HeadRefName'
WHERE external_service_type = 'github';

UPDATE changesets SET external_branch = metadata -> 'fromRef' -> 'id'
WHERE external_service_type = 'bitbucketServer';

UPDATE changesets SET external_branch = replace(changesets.external_branch, 'refs/heads/','')
WHERE external_service_type = 'bitbucketServer';

COMMIT;
