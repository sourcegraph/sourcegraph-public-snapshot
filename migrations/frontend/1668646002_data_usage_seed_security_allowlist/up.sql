INSERT INTO security_event_logs_export_allowlist (event_name)
 VALUES ('AccountCreated'),
        ('AccountDeleted'),
        ('AccountNuked'),
        ('RoleChangeGranted'),
        ('SignOutSucceeded'),
        ('PasswordChanged')
 ON CONFLICT DO NOTHING;
