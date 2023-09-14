package com.sourcegraph.cody.statusbar

import com.intellij.openapi.project.Project
import com.intellij.openapi.util.Disposer
import com.intellij.openapi.wm.StatusBarWidget
import com.intellij.openapi.wm.impl.status.widget.StatusBarEditorBasedWidgetFactory

class CodyWidgetFactory : StatusBarEditorBasedWidgetFactory() {
  override fun getId(): String = ID

  override fun getDisplayName(): String = "Cody"

  override fun createWidget(project: Project): StatusBarWidget = CodyStatusBarWidget(project)

  override fun disposeWidget(widget: StatusBarWidget) {
    Disposer.dispose(widget)
  }

  companion object {
    const val ID = "cody.statusBarWidget"
  }
}
