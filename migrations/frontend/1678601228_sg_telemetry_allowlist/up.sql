-- This migration was generated by the command `sg telemetry add`
INSERT INTO event_logs_export_allowlist (event_name) VALUES (UNNEST('{AccountCreated,AccountDeleted,AccountNuked,SignOutAttempted,SignOutFailed,SignOutSucceeded,PasswordResetRequested,PasswordChanged,RoleChangeGranted,ViewBlob,SearchSubmitted,"searchResults:owernshipMailto:clicked","searchResults:ownershipUsers:clicked","searchResults:ownershipTeams:clicked","own:ingestedCodeownersFile:added","own:ingestedCodeownersFile:updated","own:ingestedCodeownersFile:deleted","filePage:ownershipPanel:viewOwnerDetail:clicked","searchResults:ownershipCsv:exported","repoPage:ownershipPage:viewed","repoPage:ownershipPage:clicked",FileHasOwnersSearch,SelectFileOwnersSearch,OwnershipPanelOpened}'::TEXT[])) ON CONFLICT DO NOTHING;
