package com.sourcegraph.cody.statusbar

import com.intellij.openapi.actionSystem.DataContext
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.popup.ListPopup
import com.intellij.openapi.vfs.VirtualFile
import com.intellij.openapi.wm.StatusBarWidget
import com.intellij.openapi.wm.WindowManager
import com.intellij.openapi.wm.impl.status.EditorBasedStatusBarPopup

class CodyStatusBarWidget(project: Project) : EditorBasedStatusBarPopup(project, false) {
  override fun ID(): String = CodyWidgetFactory.ID

  override fun getWidgetState(file: VirtualFile?): WidgetState {
    val currentStatus = CodyAutocompleteStatusService.getCurrentStatus()
    if (currentStatus == CodyAutocompleteStatus.CodyDisabled) {
      return WidgetState.HIDDEN
    }
    val state = WidgetState(currentStatus.presentableText, "", true)
    state.icon = currentStatus.icon
    return state
  }

  override fun createPopup(context: DataContext?): ListPopup? {
    return null
  }

  override fun createInstance(project: Project): StatusBarWidget {
    return CodyStatusBarWidget(project)
  }

  companion object {

    fun update(project: Project) {
      val widget: CodyStatusBarWidget? = findWidget(project)
      widget?.update { widget.myStatusBar.updateWidget(CodyWidgetFactory.ID) }
    }

    private fun findWidget(project: Project): CodyStatusBarWidget? {
      val widget =
          WindowManager.getInstance().getStatusBar(project)?.getWidget(CodyWidgetFactory.ID)
      return if (widget is CodyStatusBarWidget) widget else null
    }
  }
}
