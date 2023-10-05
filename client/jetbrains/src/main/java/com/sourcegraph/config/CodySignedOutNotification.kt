package com.sourcegraph.config

import com.intellij.notification.Notification
import com.intellij.notification.NotificationType
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.options.ShowSettingsUtil
import com.intellij.openapi.project.DumbAwareAction
import com.intellij.openapi.project.Project
import com.sourcegraph.Icons
import com.sourcegraph.cody.config.CodyApplicationSettings
import com.sourcegraph.cody.config.ui.AccountConfigurable
import javax.annotation.concurrent.GuardedBy

object CodySignedOutNotification {
  @GuardedBy("this") private var current: Notification? = null

  fun show(project: Project) {
    synchronized(this) {
      if (current != null ||
          CodyApplicationSettings.getInstance().isNotLoggedInNotificationDismissed)
          return
      val notification =
          Notification(
              "Sourcegraph errors",
              "Sourcegraph",
              "You are no longer logged in",
              NotificationType.INFORMATION)

      notification.whenExpired { synchronized(this) { current = null } }
      notification.addAction(
          object : DumbAwareAction("Manage Accounts") {
            override fun actionPerformed(e: AnActionEvent) {
              ShowSettingsUtil.getInstance()
                  .showSettingsDialog(project, AccountConfigurable::class.java)
            }
          })
      notification.addAction(
          object : DumbAwareAction("Never Show Again") {
            override fun actionPerformed(e: AnActionEvent) {
              notification.expire()
              CodyApplicationSettings.getInstance().isNotLoggedInNotificationDismissed = true
            }
          })
      notification.setIcon(Icons.CodyLogo)
      notification.notify(project)
      current = notification
    }
  }

  fun expire(): Unit = synchronized(this) { current?.expire() }
}
