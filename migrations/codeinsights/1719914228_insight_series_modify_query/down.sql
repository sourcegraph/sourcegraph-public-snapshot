UPDATE insight_series SET query = query_old;
ALTER TABLE insight_series DROP COLUMN IF EXISTS query_old;
