-- This migration was generated by the command `sg telemetry remove`
DELETE FROM event_logs_export_allowlist WHERE event_name IN (SELECT * FROM UNNEST('{SearchSubmitted,AccessRequestApproved,AccessRequestRejected}'::TEXT[]));