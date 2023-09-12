package com.sourcegraph.cody.config

import com.intellij.openapi.actionSystem.ActionGroup
import com.intellij.openapi.actionSystem.ActionManager
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.JBPopupMenu
import com.intellij.ui.awt.RelativePoint
import com.intellij.util.containers.orNull
import com.sourcegraph.cody.auth.ui.AccountsListModel
import com.sourcegraph.cody.auth.ui.AccountsListModelBase
import com.sourcegraph.cody.localapp.LocalAppManager
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
          val token = newCredentials[account] ?: getOldToken(account)
          CodyAuthenticationManager.getInstance()
              .login(
                  project,
                  parentComponent,
                  CodyLoginRequest(
                      login = account.name,
                      server = account.server,
                      token = token,
                      customRequestHeaders = account.server.customRequestHeaders,
                      title = "Edit Sourcegraph Account",
                      loginButtonText = "Save account",
                      isServerEditable = true,
                  ))
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
    account.server.url = authData.server.url
    account.server.customRequestHeaders = authData.server.customRequestHeaders
    newCredentials[account] = authData.token
    notifyCredentialsChanged(account)
  }

  private fun getOldToken(account: CodyAccount) =
      CodyAuthenticationManager.getInstance().getTokenForAccount(account)

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
      accountsListModel.items.none { it.name == login && it.server.url == server.url }
}
