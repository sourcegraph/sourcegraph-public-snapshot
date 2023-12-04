ALTER TABLE ONLY exhaustive_search_jobs
    DROP CONSTRAINT IF EXISTS exhaustive_search_jobs_query_initiator_id_key,
    ADD CONSTRAINT exhaustive_search_jobs_query_initiator_id_key UNIQUE (query, initiator_id);
