BEGIN;

-- The "sg_service" role does not own anything, so it's safe to drop.
DROP OWNED BY sg_service;
DROP ROLE IF EXISTS sg_service;

COMMIT;
