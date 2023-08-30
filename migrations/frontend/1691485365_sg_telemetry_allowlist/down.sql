-- This migration was generated by the command `sg telemetry add`
DELETE FROM event_logs_export_allowlist WHERE event_name IN (SELECT * FROM UNNEST('{web:codyChat:pageViewed,web:codyChat:submit,web:codyChat:edit,web:codyChat:initialized,web:codyChat:editorWidgetViewed,web:codyChat:historyCleared,web:codyChat:historyItemDeleted,web:codyChat:scopeRepoAdded,web:codyChat:scopeRepoRemoved,web:codyChat:scopeReset,web:codyChat:inferredRepoEnabled,web:codyChat:inferredRepoDisabled,web:codyChat:inferredFileEnabled,web:codyChat:inferredFileDisabled,ViewGetCody,web:codyEditorWidget:viewed,web:codySidebar:chatOpened,CodySignup,web:codyChat:downloadVSCodeCTA,web:codyChat:tryOnPublicCodeCTA,CodyClickViewEditorExtensions,VSCodeInstall,VSCodeMarketplace,TryCodyWeb,TryCodyWebOnboardingDisplayed,CodySignUpInitiated,SpeakToACodyEngineerCTA,SignupInitiated,JoinIDEWaitlist,DownloadIDE,DownloadApp}'::TEXT[]));