package com.sourcegraph.cody.config

data class SettingsModel(
    var defaultBranchName: String = "",
    var remoteUrlReplacements: String = "",
    var customRequestHeaders: String = "",
    var isCodyEnabled: Boolean = false,
    var isCodyAutocompleteEnabled: Boolean = false,
    var isCodyDebugEnabled: Boolean = false,
    var isCodyVerboseDebugEnabled: Boolean = false,
    var isUrlNotificationDismissed: Boolean = false,
)
