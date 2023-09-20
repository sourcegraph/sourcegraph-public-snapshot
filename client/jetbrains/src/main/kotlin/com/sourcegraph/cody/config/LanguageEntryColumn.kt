package com.sourcegraph.cody.config

import com.intellij.util.ui.ColumnInfo
import java.util.*

class LanguageEntryColumn : ColumnInfo<LanguageEntry, String>("Enabled Languages") {
  override fun valueOf(languageEntry: LanguageEntry): String {
    return languageEntry.language.displayName
  }

  override fun getComparator(): Comparator<LanguageEntry>? {
    return Comparator.comparing { le: LanguageEntry -> le.language.displayName.lowercase() }
  }
}
