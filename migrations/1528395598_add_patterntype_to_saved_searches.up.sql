BEGIN;
UPDATE TABLE saved_searches
SET description = CONCAT(description, ' patternType:regexp')
COMMIT;
