package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.ui.AccountsListModel
import com.intellij.collaboration.auth.ui.AccountsListModelBase
import com.intellij.openapi.actionSystem.ActionGroup
import com.intellij.openapi.actionSystem.ActionManager
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.JBPopupMenu
import com.intellij.ui.awt.RelativePoint
import com.intellij.util.containers.orNull
import com.sourcegraph.cody.localapp.LocalAppManager
import java.util.UUID
import javax.swing.JComponent

class CodyAccountListModel(private val project: Project) :
    AccountsListModelBase<CodyAccount, String>(),
    AccountsListModel.WithDefault<CodyAccount, String>,
    CodyAccountsHost {

  private val actionManager = ActionManager.getInstance()

  override var defaultAccount: CodyAccount? = null

  override fun addAccount(parentComponent: JComponent, point: RelativePoint?) {
    val group = actionManager.getAction("Cody.Accounts.AddAccount") as ActionGroup
    val popup = actionManager.createActionPopupMenu("AddCodyAccountWithToken", group)

    val actualPoint = point ?: RelativePoint.getCenterOf(parentComponent)
    popup.setTargetComponent(parentComponent)
    JBPopupMenu.showAt(actualPoint, popup.component)
  }

  override fun editAccount(parentComponent: JComponent, account: CodyAccount) {
    val authData =
        if (!account.isCodyApp()) {
          CodyAuthenticationManager.getInstance()
              .login(
                  project,
                  parentComponent,
                  CodyLoginRequest(server = account.server, login = account.name))
        } else {
          val localAppAccessToken = LocalAppManager.getLocalAppAccessToken().orNull()
          if (localAppAccessToken != null) {
            CodyAuthData(account, LocalAppManager.LOCAL_APP_ID, localAppAccessToken)
          } else {
            null
          }
        }

    if (authData == null) return

    account.name = authData.login
    newCredentials[account] = authData.token
    notifyCredentialsChanged(account)
  }

  override fun addAccount(server: SourcegraphServerPath, login: String, token: String) {
    val account = CodyAccount.create(login, server)
    addAccount(account, token)
  }

  override fun addAccount(account: CodyAccount, token: String) {
    accountsListModel.add(account)
    newCredentials[account] = token
    notifyCredentialsChanged(account)
  }

  override fun isAccountUnique(login: String, server: SourcegraphServerPath): Boolean =
      accountsListModel.items.none { it.name == login && it.server == server }
}
