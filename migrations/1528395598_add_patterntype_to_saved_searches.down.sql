BEGIN;
UPDATE TABLE saved_searches
SET description = TRIM(TRAILING ' patternType:regexp' FROM description)
COMMIT;
