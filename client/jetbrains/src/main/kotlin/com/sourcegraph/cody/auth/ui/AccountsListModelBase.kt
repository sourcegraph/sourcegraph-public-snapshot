package com.sourcegraph.cody.auth.ui

import com.intellij.ui.CollectionListModel
import com.sourcegraph.cody.auth.Account
import com.sourcegraph.cody.auth.SingleValueModel
import java.util.concurrent.CopyOnWriteArrayList

abstract class AccountsListModelBase<A : Account, Cred> : AccountsListModel<A, Cred> {
  override var accounts: Set<A>
    get() = accountsListModel.items.toSet()
    set(value) {
      accountsListModel.removeAll()
      accountsListModel.add(value.toList())
    }
  override var selectedAccount: A? = null
  override val newCredentials = mutableMapOf<A, Cred>()

  override val accountsListModel = CollectionListModel<A>()
  override val busyStateModel = SingleValueModel(false)

  private val credentialsChangesListeners = CopyOnWriteArrayList<(A) -> Unit>()

  override fun clearNewCredentials() = newCredentials.clear()

  protected fun notifyCredentialsChanged(account: A) {
    credentialsChangesListeners.forEach { it(account) }
    accountsListModel.contentsChanged(account)
  }

  final override fun addCredentialsChangeListener(listener: (A) -> Unit) {
    credentialsChangesListeners.add(listener)
  }
}
