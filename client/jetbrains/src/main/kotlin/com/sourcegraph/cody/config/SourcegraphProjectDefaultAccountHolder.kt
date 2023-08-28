package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.PersistentDefaultAccountHolder
import com.intellij.notification.Notifications
import com.intellij.openapi.components.State
import com.intellij.openapi.components.Storage
import com.intellij.openapi.components.StoragePathMacros
import com.intellij.openapi.components.service
import com.intellij.openapi.project.Project
import com.sourcegraph.common.ErrorNotification
import com.sourcegraph.config.OpenPluginSettingsAction

@State(
    name = "SourcegraphDefaultAccount",
    storages = [Storage(StoragePathMacros.WORKSPACE_FILE)],
    reportStatistic = false)
class SourcegraphProjectDefaultAccountHolder(project: Project) :
    PersistentDefaultAccountHolder<SourcegraphAccount>(project) {
  override fun accountManager() = service<SourcegraphAccountManager>()

  override fun notifyDefaultAccountMissing() {
    val notification = ErrorNotification.create("Default Sourcegraph account was not found")
    notification.addAction(OpenPluginSettingsAction("Configure Account"))
    Notifications.Bus.notify(notification, project)
  }
}
