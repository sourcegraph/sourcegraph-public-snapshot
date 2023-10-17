package com.sourcegraph.cody.statusbar

import com.intellij.ide.actions.AboutAction
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
    if (CodyAutocompleteStatusService.getCurrentStatus() ==
        CodyAutocompleteStatus.CodyAgentNotRunning)
        return listOf(
                OpenLogAction(),
                AboutAction().apply {
                  templatePresentation.text = "Open About To Troubleshoot Issue"
                },
                ReportCodyBugAction())
            .toTypedArray()

    return listOfNotNull(
            CodyEnableAutocompleteAction(),
            CodyDisableAutocompleteAction(),
            CodyEnableLanguageForAutocompleteAction(),
            CodyDisableLanguageForAutocompleteAction(),
            CodyManageAccountsAction(),
            CodyOpenSettingsAction(),
        )
        .toTypedArray()
  }
}
