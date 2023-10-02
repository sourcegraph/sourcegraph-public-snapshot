package com.sourcegraph.cody.statusbar

import com.intellij.idea.ActionsBundle
import com.intellij.internal.OpenLogAction
import com.intellij.openapi.actionSystem.AnAction
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.actionSystem.DefaultActionGroup
import com.sourcegraph.config.ConfigUtil

class CodyStatusBarActionGroup : DefaultActionGroup() {
  override fun update(e: AnActionEvent) {
    super.update(e)
    e.presentation.isVisible = ConfigUtil.isCodyEnabled()
  }

  override fun getChildren(e: AnActionEvent?): Array<AnAction> {
    return listOfNotNull(
            CodyEnableAutocompleteAction(),
            CodyDisableAutocompleteAction(),
            CodyEnableLanguageForAutocompleteAction(),
            CodyDisableLanguageForAutocompleteAction(),
            CodyManageAccountsAction(),
            CodyOpenSettingsAction(),
            if (CodyAutocompleteStatusService.getCurrentStatus() ==
                CodyAutocompleteStatus.CodyAgentNotRunning)
                OpenLogAction().apply {
                  templatePresentation.text = ActionsBundle.message("action.OpenLog.text")
                }
            else null)
        .toTypedArray()
  }
}
