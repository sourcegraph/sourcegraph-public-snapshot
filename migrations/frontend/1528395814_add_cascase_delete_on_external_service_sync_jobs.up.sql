BEGIN;

SET CONSTRAINTS ALL DEFERRED;

ALTER TABLE external_service_sync_jobs
    DROP CONSTRAINT external_services_id_fk,
    ADD CONSTRAINT external_services_id_fk
        FOREIGN KEY (external_service_id)
            REFERENCES external_services (id)
            ON DELETE CASCADE;

COMMIT;
