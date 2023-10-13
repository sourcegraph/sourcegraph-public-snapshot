package com.sourcegraph.cody.statusbar

import com.intellij.ide.BrowserUtil
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.project.DumbAwareAction

class ReportCodyBugAction : DumbAwareAction("Open GitHub To Report Cody Issue") {
  override fun actionPerformed(p0: AnActionEvent) {
    BrowserUtil.open("https://github.com/sourcegraph/sourcegraph/issues/new?template=jetbrains.md")
  }
}
