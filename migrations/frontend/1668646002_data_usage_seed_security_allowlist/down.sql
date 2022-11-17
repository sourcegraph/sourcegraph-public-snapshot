DELETE
FROM security_event_logs_export_allowlist
WHERE event_name IN ('AccountCreated',
                     'AccountDeleted',
                     'AccountNuked',
                     'RoleChangeGranted',
                     'SignOutSucceeded',
                     'PasswordChanged');
