package com.sourcegraph.cody.config

import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.components.service
import com.intellij.openapi.project.DumbAwareAction
import com.sourcegraph.cody.localapp.LocalAppManager

class AddCodyAccountAction : DumbAwareAction() {
  override fun update(e: AnActionEvent) {
    val localAppInstalled = LocalAppManager.isLocalAppInstalled()
    val accountManager = service<SourcegraphAccountManager>()
    val codyAccountAlreadyAdded = accountManager.accounts.any { it.isCodyApp() }
    e.presentation.isEnabledAndVisible =
        e.getData(SourcegraphAccountsHost.KEY) != null &&
            localAppInstalled &&
            !codyAccountAlreadyAdded
  }

  override fun actionPerformed(e: AnActionEvent) {
    val accountsHost = e.getData(SourcegraphAccountsHost.KEY)!!
    val token = LocalAppManager.getLocalAppAccessToken()
    val account =
        SourcegraphAccount.create(
            LocalAppManager.LOCAL_APP_ID,
            SourcegraphServerPath(LocalAppManager.getLocalAppUrl()),
            LocalAppManager.LOCAL_APP_ID)
    if (accountsHost.isAccountUnique(account.name, account.server)) {
      token.ifPresent { accountsHost.addAccount(account, it) }
    }
  }
}
