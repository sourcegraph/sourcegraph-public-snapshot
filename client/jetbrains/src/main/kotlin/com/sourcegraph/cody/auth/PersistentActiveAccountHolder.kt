package com.sourcegraph.cody.auth

import com.intellij.openapi.Disposable
import com.intellij.openapi.components.PersistentStateComponent
import com.intellij.openapi.project.Project

/**
 * Stores active account for project To
 * register - [@State(name = SERVICE_NAME_HERE, storages = [Storage(StoragePathMacros.WORKSPACE_FILE)],
 * reportStatistic = false)]
 *
 * @param A - account type
 */
abstract class PersistentActiveAccountHolder<A : Account>(protected val project: Project) :
    PersistentStateComponent<PersistentActiveAccountHolder.AccountState>, Disposable {

  var account: A? = null

  private val accountManager: AccountManager<A, *>
    get() = accountManager()

  init {
    accountManager.addListener(
        this,
        object : AccountsListener<A> {
          override fun onAccountListChanged(old: Collection<A>, new: Collection<A>) {
            if (!new.contains(account)) account = null
          }
        })
  }

  override fun getState(): AccountState {
    return AccountState().apply { activeAccountId = account?.id }
  }

  override fun loadState(state: AccountState) {
    account =
        state.activeAccountId?.let { id ->
          accountManager.accounts
              .find { it.id == id }
              .also { if (it == null) notifyActiveAccountMissing() }
        }
  }

  protected abstract fun accountManager(): AccountManager<A, *>

  protected abstract fun notifyActiveAccountMissing()

  override fun dispose() {}

  class AccountState {
    var activeAccountId: String? = null
  }
}
