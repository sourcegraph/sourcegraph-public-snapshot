// NOTE(naman): Remember to add events to allow list: https://docs.sourcegraph.com/dev/background-information/data-usage-pipeline#allow-list
export const enum EventName {
    CODY_CHAT_PAGE_VIEWED = 'web:codyChat:pageViewed',
    CODY_CHAT_SUBMIT = 'web:codyChat:submit',
    CODY_CHAT_EDIT = 'web:codyChat:edit',
    CODY_CHAT_INITIALIZED = 'web:codyChat:initialized',
    CODY_CHAT_EDITOR_WIDGET_VIEWED = 'web:codyChat:editorWidgetViewed',
    CODY_CHAT_HISTORY_CLEARED = 'web:codyChat:historyCleared',
    CODY_CHAT_HISTORY_ITEM_DELETED = 'web:codyChat:historyItemDeleted',

    CODY_CHAT_SCOPE_REPO_ADDED = 'web:codyChat:scopeRepoAdded',
    CODY_CHAT_SCOPE_REPO_REMOVED = 'web:codyChat:scopeRepoRemoved',
    CODY_CHAT_SCOPE_RESET = 'web:codyChat:scopeReset',
    CODY_CHAT_SCOPE_INFERRED_REPO_ENABLED = 'web:codyChat:inferredRepoEnabled',
    CODY_CHAT_SCOPE_INFERRED_REPO_DISABLED = 'web:codyChat:inferredRepoDisabled',
    CODY_CHAT_SCOPE_INFERRED_FILE_ENABLED = 'web:codyChat:inferredFileEnabled',
    CODY_CHAT_SCOPE_INFERRED_FILE_DISABLED = 'web:codyChat:inferredFileDisabled',
    VIEW_GET_CODY = 'GetCody',

    CODY_EDITOR_WIDGET_VIEWED = 'web:codyEditorWidget:viewed',
    CODY_SIDEBAR_CHAT_OPENED = 'web:codySidebar:chatOpened',
    CODY_SIGNUP = 'CodySignup',
    CODY_CHAT_DOWNLOAD_VSCODE = 'web:codyChat:downloadVSCodeCTA',
    CODY_CHAT_GET_EDITOR_EXTENSION = 'web:codyChat:getEditorExtensionCTA',
    CODY_CHAT_TRY_ON_PUBLIC_CODE = 'web:codyChat:tryOnPublicCodeCTA',
    CODY_CTA = 'ClickedOnCodyCTA',
    VIEW_EDITOR_EXTENSIONS = 'CodyClickViewEditorExtensions',
    TRY_CODY_VSCODE = 'VSCodeInstall',
    TRY_CODY_MARKETPLACE = 'VSCodeMarketplace',
    TRY_CODY_WEB = 'TryCodyWeb',
    TRY_CODY_WEB_ONBOARDING_DISPLAYED = 'TryCodyWebOnboardingDisplayed',
    TRY_CODY_SIGNUP_INITIATED = 'CodySignUpInitiated',
    SPEAK_TO_AN_ENGINEER_CTA = 'SpeakToACodyEngineerCTA',
    SIGNUP_INITIATED = 'SignupInitiated',

    JOIN_IDE_WAITLIST = 'JoinIDEWaitlist',
    DOWNLOAD_IDE = 'DownloadIDE',
    DOWNLOAD_APP = 'DownloadApp',
}

export const enum EventLocation {
    NAV_BAR = 'NavBar',
    CHAT_RESPONSE = 'ChatResponse',
}
