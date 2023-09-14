package com.sourcegraph.cody.config

import com.intellij.openapi.components.State
import com.intellij.openapi.components.Storage
import com.intellij.openapi.components.StoragePathMacros
import com.intellij.openapi.components.service
import com.intellij.openapi.project.Project
import com.sourcegraph.cody.auth.PersistentActiveAccountHolder

@State(
    name = "CodyActiveAccount",
    storages = [Storage(StoragePathMacros.WORKSPACE_FILE)],
    reportStatistic = false)
class CodyProjectActiveAccountHolder(project: Project) :
    PersistentActiveAccountHolder<CodyAccount>(project) {

  override fun accountManager() = service<CodyAccountManager>()

  override fun notifyActiveAccountMissing() {}
}
