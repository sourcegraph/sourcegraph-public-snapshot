package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.AccountsListener
import com.intellij.collaboration.auth.PersistentDefaultAccountHolder
import com.intellij.openapi.components.State
import com.intellij.openapi.components.Storage
import com.intellij.openapi.components.StoragePathMacros
import com.intellij.openapi.components.service
import com.intellij.openapi.project.Project

@State(
    name = "SourcegraphDefaultAccount",
    storages = [Storage(StoragePathMacros.WORKSPACE_FILE)],
    reportStatistic = false)
class SourcegraphProjectDefaultAccountHolder(project: Project) :
    PersistentDefaultAccountHolder<SourcegraphAccount>(project) {

  init {
    accountManager()
        .addListener(
            this,
            object : AccountsListener<SourcegraphAccount> {
              override fun onAccountListChanged(
                  old: Collection<SourcegraphAccount>,
                  new: Collection<SourcegraphAccount>
              ) {
                if (account == null) {
                  account = new.firstOrNull { it.isCodyApp() } ?: new.firstOrNull()
                }
              }
            })
  }
  override fun accountManager() = service<SourcegraphAccountManager>()

  override fun notifyDefaultAccountMissing() {}
}
