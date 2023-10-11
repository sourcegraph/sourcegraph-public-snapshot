package com.sourcegraph.cody.statusbar

import com.intellij.idea.LoggerFactory
import com.intellij.notification.Notification
import com.intellij.notification.NotificationType
import com.intellij.notification.Notifications.Bus
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.fileEditor.FileEditorManager
import com.intellij.openapi.project.DumbAwareAction
import com.intellij.openapi.vfs.LocalFileSystem
import com.intellij.openapi.vfs.VfsUtil

class OpenLogAction() : DumbAwareAction("Open Log To Troubleshoot Issue") {

  override fun actionPerformed(e: AnActionEvent) {
    val project = e.project
    if (project != null) {
      val file =
          LocalFileSystem.getInstance().refreshAndFindFileByNioFile(LoggerFactory.getLogFilePath())
      if (file != null) {
        VfsUtil.markDirtyAndRefresh(true, false, false, *arrayOf(file))
        FileEditorManager.getInstance(project).openFile(file, true)
      } else {
        val title = "Cannot find '" + LoggerFactory.getLogFilePath() + "'"
        Bus.notify(Notification("System Messages", title, "", NotificationType.INFORMATION))
      }
    }
  }
}
