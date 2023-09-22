package com.sourcegraph.cody.config.ui.lang

import com.intellij.lang.Language
import com.sourcegraph.config.ConfigUtil
import java.util.stream.Collectors

class LanguageEntry(val language: Language, var isBlacklisted: Boolean) {
  companion object {
    val ANY = LanguageEntry(Language.ANY, false)

    fun getRegisteredLanguageEntries(): List<LanguageEntry> {
      val blacklistedLanguageIds: List<String> = ConfigUtil.getBlacklistedAutocompleteLanguageIds()
      return Language.getRegisteredLanguages()
          .stream()
          .filter { l ->
            l != Language.ANY
          } // skip ANY, since it doesn't correspond to any actual language
          .map { l -> LanguageEntry(l!!, blacklistedLanguageIds.contains(l.id)) }
          .collect(Collectors.toList())
          .sortedBy { le -> le.language.displayName }
    }
  }
}
