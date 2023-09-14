package com.sourcegraph.cody.auth.ui

import com.intellij.ui.awt.RelativePoint
import com.sourcegraph.cody.auth.Account
import com.sourcegraph.cody.auth.SingleValueModel
import javax.swing.JComponent
import javax.swing.ListModel

interface AccountsListModel<A: Account, Cred> {
  var accounts: Set<A>
  var selectedAccount: A?
  val newCredentials: Map<A, Cred>

  val accountsListModel: ListModel<A>
  val busyStateModel: SingleValueModel<Boolean>

  fun addAccount(parentComponent: JComponent, point: RelativePoint? = null)
  fun editAccount(parentComponent: JComponent, account: A)
  fun clearNewCredentials()

  fun addCredentialsChangeListener(listener: (A) -> Unit)

  interface WithActive<A: Account, Cred>: AccountsListModel<A, Cred> {
    var activeAccount: A?
  }
}
