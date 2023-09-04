package com.sourcegraph.cody.config

import com.intellij.openapi.components.service
import com.intellij.openapi.project.Project
import com.intellij.util.AuthData
import com.intellij.util.concurrency.annotations.RequiresEdt
import com.sourcegraph.cody.localapp.LocalAppManager
import java.awt.Component
import org.jetbrains.annotations.CalledInAny

internal class SourcegraphAuthData(val account: SourcegraphAccount, login: String, token: String) :
    AuthData(login, token) {
  val server: SourcegraphServerPath
    get() = account.server
  val token: String
    get() = password!!
}

/** Entry point for interactions with Sourcegraph authentication subsystem */
class SourcegraphAuthenticationManager internal constructor() {
  private val accountManager: SourcegraphAccountManager
    get() = service()

  @CalledInAny fun getAccounts(): Set<SourcegraphAccount> = accountManager.accounts

  @CalledInAny
  internal fun getTokenForAccount(account: SourcegraphAccount): String? =
      accountManager.findCredentials(account)

  internal fun isAccountUnique(name: String, server: SourcegraphServerPath) =
      accountManager.accounts.none { it.name == name && it.server == server }

  @RequiresEdt
  internal fun login(
      project: Project?,
      parentComponent: Component?,
      request: SourcegraphLoginRequest
  ): SourcegraphAuthData? = request.loginWithToken(project, parentComponent)

  @RequiresEdt
  internal fun updateAccountToken(account: SourcegraphAccount, newToken: String) =
      accountManager.updateAccount(account, newToken)

  fun getDefaultAccount(project: Project): SourcegraphAccount? =
      project.service<SourcegraphProjectDefaultAccountHolder>().account

  fun setDefaultAccount(project: Project, account: SourcegraphAccount?) {
    project.service<SourcegraphProjectDefaultAccountHolder>().account = account
  }

  fun getDefaultAccountType(project: Project): AccountType {
    return getDefaultAccount(project)?.getAccountType()
        ?: if (LocalAppManager.isLocalAppInstalled() && LocalAppManager.isPlatformSupported())
            AccountType.LOCAL_APP
        else AccountType.DOTCOM
  }

  companion object {
    @JvmStatic fun getInstance(): SourcegraphAuthenticationManager = service()
  }
}
