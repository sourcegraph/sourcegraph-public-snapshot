-- This will reset the state of the insight definition tables to empty so that the OOB migration can fully
-- replicate and migrate all of the insights from settings without any duplication.

TRUNCATE insight_view CASCADE;
TRUNCATE insight_series CASCADE;
TRUNCATE dashboard CASCADE;
