package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.AccountsRepository
import com.intellij.openapi.components.PersistentStateComponent
import com.intellij.openapi.components.SettingsCategory
import com.intellij.openapi.components.State
import com.intellij.openapi.components.Storage
import com.sourcegraph.cody.localapp.LocalAppManager

@State(
    name = "SourcegraphAccounts",
    storages =
        [
            Storage(value = "sourcegraph_accounts.xml"),
        ],
    reportStatistic = false,
    category = SettingsCategory.TOOLS)
class SourcegraphPersisentAccounts :
    AccountsRepository<SourcegraphAccount>, PersistentStateComponent<Array<SourcegraphAccount>> {
  private var state = emptyArray<SourcegraphAccount>()

  override var accounts: Set<SourcegraphAccount>
    get() = state.toSet()
    set(value) {
      state = value.toTypedArray()
    }

  override fun getState(): Array<SourcegraphAccount> = state

  override fun loadState(state: Array<SourcegraphAccount>) {
    var finalState = state
    if (state.none { it.id == LocalAppManager.LOCAL_APP_ID }) {
      val localAppInstalled = LocalAppManager.isLocalAppInstalled()
      if (localAppInstalled) {
        finalState =
            state +
                SourcegraphAccount(
                    LocalAppManager.LOCAL_APP_ID,
                    SourcegraphServerPath(LocalAppManager.getLocalAppUrl()))
      }
    }
    this.state = finalState
  }
}
