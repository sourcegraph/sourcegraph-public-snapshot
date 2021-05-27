BEGIN;

UPDATE lsif_uploads SET state = 'deleted' WHERE state = 'DELETED';

COMMIT;
