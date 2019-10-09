BEGIN;

UPDATE saved_searches SET query=CONCAT(query, ' patternType:regexp');

COMMIT;
