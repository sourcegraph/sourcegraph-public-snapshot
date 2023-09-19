package com.sourcegraph.cody.statusbar

import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.project.DumbAwareAction
import com.sourcegraph.cody.autocomplete.CodyAutocompleteManager
import com.sourcegraph.cody.config.CodyApplicationSettings
import com.sourcegraph.config.ConfigUtil
import com.sourcegraph.utils.CodyEditorUtil
import com.sourcegraph.utils.CodyLanguageUtil

class CodyDisableLanguageForAutocompleteAction : DumbAwareAction() {
  override fun actionPerformed(e: AnActionEvent) {
    val applicationSettings = CodyApplicationSettings.getInstance()
    CodyEditorUtil.getLanguageForFocusedEditor(e)?.id?.let { languageId ->
      applicationSettings.blacklistedLanguageIds =
          applicationSettings.blacklistedLanguageIds.plus(languageId)
      CodyAutocompleteManager.getInstance().clearAutocompleteSuggestionsForLanguageId(languageId)
    }
  }

  override fun update(e: AnActionEvent) {
    super.update(e)
    val languageForFocusedEditor = CodyEditorUtil.getLanguageForFocusedEditor(e)
    val isLanguageBlacklisted =
        languageForFocusedEditor?.let { CodyLanguageUtil.isLanguageBlacklisted(it) } ?: false
    val languageName = languageForFocusedEditor?.displayName ?: ""
    e.presentation.isEnabledAndVisible =
        languageForFocusedEditor != null &&
            ConfigUtil.isCodyEnabled() &&
            ConfigUtil.isCodyAutocompleteEnabled() &&
            !isLanguageBlacklisted
    e.presentation.text = "Disable Cody Autocomplete for $languageName"
  }
}
