package com.sourcegraph.cody.statusbar

import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.project.DumbAwareAction
import com.sourcegraph.cody.autocomplete.CodyAutocompleteManager
import com.sourcegraph.config.ConfigUtil

class CodyDisableAutocompleteAction : DumbAwareAction() {
    override fun actionPerformed(e: AnActionEvent) {
        ConfigUtil.setCodyAutocompleteEnabled(false)
        CodyAutocompleteManager.getInstance().clearAutocompleteSuggestionsForAllProjects()
    }

    override fun update(e: AnActionEvent) {
        super.update(e)
        e.presentation.isEnabledAndVisible = ConfigUtil.isCodyEnabled() && ConfigUtil.isCodyAutocompleteEnabled()
    }
}
