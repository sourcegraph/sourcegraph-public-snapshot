BEGIN;

UPDATE repo SET external_id = NULL
WHERE LOWER(external_service_type) = 'bitbucketserver';

COMMIT;
