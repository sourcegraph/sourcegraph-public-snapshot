package com.sourcegraph.cody.statusbar

import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.options.ShowSettingsUtil
import com.intellij.openapi.project.DumbAwareAction
import com.sourcegraph.cody.config.ui.CodyConfigurable

class CodyOpenSettingsAction : DumbAwareAction() {
  override fun actionPerformed(e: AnActionEvent) {
    ShowSettingsUtil.getInstance().showSettingsDialog(e.project, CodyConfigurable::class.java)
  }
}
