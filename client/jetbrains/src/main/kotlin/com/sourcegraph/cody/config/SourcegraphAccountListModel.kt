package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.ui.AccountsListModel
import com.intellij.collaboration.auth.ui.AccountsListModelBase
import com.intellij.openapi.actionSystem.ActionGroup
import com.intellij.openapi.actionSystem.ActionManager
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.JBPopupMenu
import com.intellij.ui.awt.RelativePoint
import java.util.UUID
import javax.swing.JComponent

class SourcegraphAccountListModel(private val project: Project) :
    AccountsListModelBase<SourcegraphAccount, String>(),
    AccountsListModel.WithDefault<SourcegraphAccount, String>,
    SourcegraphAccountsHost {

  private val actionManager = ActionManager.getInstance()

  override var defaultAccount: SourcegraphAccount? = null

  override fun addAccount(parentComponent: JComponent, point: RelativePoint?) {
    val group = actionManager.getAction("Sourcegraph.Accounts.AddAccount") as ActionGroup
    val popup = actionManager.createActionPopupMenu("AddSourcegraphAccount", group)

    val actualPoint = point ?: RelativePoint.getCenterOf(parentComponent)
    popup.setTargetComponent(parentComponent)
    JBPopupMenu.showAt(actualPoint, popup.component)
  }

  override fun editAccount(parentComponent: JComponent, account: SourcegraphAccount) {
    //        val authData = GithubAuthenticationManager.getInstance().login(
    //            project, parentComponent,
    //            GHLoginRequest(server = account.server, login = account.name)
    //        )
    //        if (authData == null) return
    //
    //        account.name = authData.login
    newCredentials[account] = UUID.randomUUID().toString()
    notifyCredentialsChanged(account)
  }

  override fun addAccount(server: SourcegraphServerPath, login: String, token: String) {
    val account = SourcegraphAccount(login, server, id = UUID.randomUUID().toString())
    accountsListModel.add(account)
    newCredentials[account] = token
    notifyCredentialsChanged(account)
  }

  override fun isAccountUnique(login: String, server: SourcegraphServerPath): Boolean =
      accountsListModel.items.none { it.name == login && it.server == server }
}
