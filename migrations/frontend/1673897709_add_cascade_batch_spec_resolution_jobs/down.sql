ALTER TABLE 
    batch_spec_resolution_jobs
DROP 
    CONSTRAINT IF EXISTS batch_spec_resolution_jobs_initiator_id_fkey;

ALTER TABLE 
    batch_spec_resolution_jobs
ADD 
    CONSTRAINT batch_spec_resolution_jobs_initiator_id_fkey 
        FOREIGN KEY (initiator_id) 
        REFERENCES users(id) 
        ON UPDATE CASCADE DEFERRABLE;