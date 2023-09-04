package com.sourcegraph.cody.config

import com.intellij.openapi.project.Project
import com.intellij.openapi.util.NlsContexts
import git4idea.DialogManager
import java.awt.Component
import java.util.UUID

internal class SourcegraphLoginRequest(
    @NlsContexts.DialogMessage val text: String? = null,
    val server: SourcegraphServerPath? = null,
    val isServerEditable: Boolean = server == null,
    val login: String? = null,
    val isCheckLoginUnique: Boolean = false,
    val token: String? = null,
    val customRequestHeaders: String? = null
)

internal fun SourcegraphLoginRequest.loginWithToken(
    project: Project?,
    parentComponent: Component?
): SourcegraphAuthData? {
  val dialog = SourcegraphTokenLoginDialog(project, parentComponent, isLoginUniqueChecker)
  configure(dialog)

  return dialog.getAuthData()
}

private val SourcegraphLoginRequest.isLoginUniqueChecker: UniqueLoginPredicate
  get() = { login, server ->
    !isCheckLoginUnique ||
        SourcegraphAuthenticationManager.getInstance().isAccountUnique(login, server)
  }

private fun SourcegraphLoginRequest.configure(dialog: BaseLoginDialog) {
  server?.let { dialog.setServer(it.toString(), isServerEditable) }
  login?.let { dialog.setLogin(it) }
  token?.let { dialog.setToken(it) }
  customRequestHeaders?.let { dialog.setCustomRequestHeaders(it) }
}

private fun BaseLoginDialog.getAuthData(): SourcegraphAuthData? {
  DialogManager.show(this)
  return if (isOK)
      SourcegraphAuthData(
          SourcegraphAccount.create(login, server, UUID.randomUUID().toString()), login, token)
  else null
}
