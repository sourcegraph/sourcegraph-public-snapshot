package com.sourcegraph.cody.config.notification

class CodySettingChangeContext(val oldCodyEnabled: Boolean,
                               val newCodyEnabled: Boolean,
                               val oldCodyAutocompleteEnabled: Boolean,
                               val newCodyAutocompleteEnabled: Boolean,
                               val oldIsCustomAutocompleteColorEnabled: Boolean,
                               val isCustomAutocompleteColorEnabled: Boolean,
                               val oldCustomAutocompleteColor: Int?,
                               val customAutocompleteColor: Int?,
                               val oldBlacklistedAutocompleteLanguageIds: List<String>,
                               val newBlacklistedAutocompleteLanguageIds: List<String>)
