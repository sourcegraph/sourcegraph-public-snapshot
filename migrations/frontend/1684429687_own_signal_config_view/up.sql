CREATE OR REPLACE VIEW own_background_jobs_config_aware AS
SELECT obj.*, osc.name AS config_name
FROM own_background_jobs obj
         JOIN own_signal_configurations osc ON obj.job_type = osc.id
WHERE osc.enabled IS TRUE;
