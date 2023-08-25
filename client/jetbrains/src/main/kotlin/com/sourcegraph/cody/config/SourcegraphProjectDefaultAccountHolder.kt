package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.AccountManager
import com.intellij.collaboration.auth.PersistentDefaultAccountHolder
import com.intellij.openapi.components.State
import com.intellij.openapi.components.Storage
import com.intellij.openapi.components.StoragePathMacros
import com.intellij.openapi.components.service
import com.intellij.openapi.project.Project

@State(name = "SourcegraphDefaultAccount", storages = [Storage(StoragePathMacros.WORKSPACE_FILE)], reportStatistic = false)
class SourcegraphProjectDefaultAccountHolder(project: Project): PersistentDefaultAccountHolder<SourcegraphAccount>(project) {
    override fun accountManager() = service<SourcegraphAccountManager>()

    override fun notifyDefaultAccountMissing() {
        TODO("Not yet implemented")
    }
}
