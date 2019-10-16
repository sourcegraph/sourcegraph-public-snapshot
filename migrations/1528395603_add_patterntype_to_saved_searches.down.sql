BEGIN;

UPDATE saved_searches SET query= TRIM(TRAILING ' patternType:regexp' FROM query);

COMMIT;
