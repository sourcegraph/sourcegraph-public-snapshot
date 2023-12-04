ALTER TABLE external_services
    DROP COLUMN IF EXISTS code_host_id;

DROP TABLE IF EXISTS code_hosts;
