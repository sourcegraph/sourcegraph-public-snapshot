DROP TRIGGER IF EXISTS trigger_lsif_uploads_update ON lsif_uploads;
DROP TRIGGER IF EXISTS trigger_lsif_uploads_delete ON lsif_uploads;

DROP FUNCTION IF EXISTS func_lsif_uploads_update;
DROP FUNCTION IF EXISTS func_lsif_uploads_delete;

DROP TABLE IF EXISTS lsif_uploads_audit_logs;

DROP INDEX IF EXISTS lsif_uploads_audit_logs_upload_id;
DROP INDEX IF EXISTS lsif_uploads_audit_logs_timestamp;
