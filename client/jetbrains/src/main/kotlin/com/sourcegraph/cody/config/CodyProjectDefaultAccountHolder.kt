package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.PersistentDefaultAccountHolder
import com.intellij.openapi.components.State
import com.intellij.openapi.components.Storage
import com.intellij.openapi.components.StoragePathMacros
import com.intellij.openapi.components.service
import com.intellij.openapi.project.Project

@State(
    name = "CodyDefaultAccount",
    storages = [Storage(StoragePathMacros.WORKSPACE_FILE)],
    reportStatistic = false)
class CodyProjectDefaultAccountHolder(project: Project) :
    PersistentDefaultAccountHolder<CodyAccount>(project) {

  override fun accountManager() = service<CodyAccountManager>()

  override fun notifyDefaultAccountMissing() {}
}
