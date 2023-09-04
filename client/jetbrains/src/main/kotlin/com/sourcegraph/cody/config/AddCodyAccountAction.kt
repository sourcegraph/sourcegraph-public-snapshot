package com.sourcegraph.cody.config

import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.project.DumbAwareAction
import com.sourcegraph.cody.localapp.LocalAppManager

class AddCodyAccountAction : DumbAwareAction() {
  override fun update(e: AnActionEvent) {
    val localAppInstalled = LocalAppManager.isLocalAppInstalled()
    val codyAuthenticationManager = CodyAuthenticationManager.getInstance()
    val codyAccountAlreadyAdded =
        codyAuthenticationManager.getAccounts().any { it.isCodyApp() }
    e.presentation.isEnabledAndVisible =
        e.getData(CodyAccountsHost.KEY) != null &&
            localAppInstalled &&
            !codyAccountAlreadyAdded
  }

  override fun actionPerformed(e: AnActionEvent) {
    val accountsHost = e.getData(CodyAccountsHost.KEY)!!
    val token = LocalAppManager.getLocalAppAccessToken()
    val account =
        CodyAccount.create(
            LocalAppManager.LOCAL_APP_ID,
            SourcegraphServerPath(LocalAppManager.getLocalAppUrl()),
            LocalAppManager.LOCAL_APP_ID)
    if (accountsHost.isAccountUnique(account.name, account.server)) {
      token.ifPresent { accountsHost.addAccount(account, it) }
    }
  }
}
