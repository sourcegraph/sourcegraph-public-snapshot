package com.sourcegraph.cody.config

import com.intellij.util.ui.ColumnInfo

class LanguageEntryColumn : ColumnInfo<LanguageEntry, String>("Enabled Languages") {
    override fun valueOf(languageEntry: LanguageEntry): String {
        return languageEntry.language.displayName
    }
}

