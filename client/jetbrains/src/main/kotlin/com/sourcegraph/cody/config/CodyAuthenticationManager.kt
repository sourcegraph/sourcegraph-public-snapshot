package com.sourcegraph.cody.config

import com.intellij.openapi.components.Service
import com.intellij.openapi.components.service
import com.intellij.openapi.project.Project
import com.intellij.util.AuthData
import com.intellij.util.concurrency.annotations.RequiresEdt
import java.awt.Component
import org.jetbrains.annotations.CalledInAny

internal class CodyAuthData(val account: CodyAccount, login: String, token: String) :
    AuthData(login, token) {
  val server: SourcegraphServerPath
    get() = account.server

  val token: String
    get() = password!!
}

/** Entry point for interactions with Sourcegraph authentication subsystem */
@Service
class CodyAuthenticationManager internal constructor() {
  private val accountManager: CodyAccountManager
    get() = service()

  @CalledInAny fun getAccounts(): Set<CodyAccount> = accountManager.accounts

  @CalledInAny
  internal fun getTokenForAccount(account: CodyAccount): String? =
      accountManager.findCredentials(account)

  internal fun isAccountUnique(name: String, server: SourcegraphServerPath) =
      accountManager.accounts.none { it.name == name && it.server.url == server.url }

  @RequiresEdt
  internal fun login(
      project: Project?,
      parentComponent: Component?,
      request: CodyLoginRequest
  ): CodyAuthData? = request.loginWithToken(project, parentComponent)

  @RequiresEdt
  internal fun updateAccountToken(account: CodyAccount, newToken: String) =
      accountManager.updateAccount(account, newToken)

  fun getActiveAccount(project: Project): CodyAccount? {
    return if (!project.isDisposed) project.service<CodyProjectActiveAccountHolder>().account
    else null
  }

  fun setActiveAccount(project: Project, account: CodyAccount?) {
    if (!project.isDisposed) project.service<CodyProjectActiveAccountHolder>().account = account
  }

  fun getActiveAccountType(project: Project): AccountType {
    return getActiveAccount(project)?.getAccountType() ?: AccountType.DOTCOM
  }

  companion object {
    @JvmStatic
    val instance: CodyAuthenticationManager
      get() = service()
  }
}
