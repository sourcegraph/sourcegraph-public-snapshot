package com.sourcegraph.cody.config

import java.awt.Color

data class SettingsModel(
    var defaultBranchName: String = "",
    var remoteUrlReplacements: String = "",
    var isCodyEnabled: Boolean = false,
    var isCodyAutocompleteEnabled: Boolean = false,
    var isCodyDebugEnabled: Boolean = false,
    var isCodyVerboseDebugEnabled: Boolean = false,
    var isUrlNotificationDismissed: Boolean = false,
    var isCustomAutocompleteColorEnabled: Boolean = false,
    var customAutocompleteColor: Color? = null,
    var blacklistedLanguageIds: List<String> = listOf(),
)
