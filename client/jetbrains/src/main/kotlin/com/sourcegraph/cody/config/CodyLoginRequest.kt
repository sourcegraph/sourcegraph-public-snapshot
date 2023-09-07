package com.sourcegraph.cody.config

import com.intellij.openapi.project.Project
import com.intellij.openapi.util.NlsContexts
import git4idea.DialogManager
import java.awt.Component
import java.util.UUID

internal class CodyLoginRequest(
    @NlsContexts.DialogMessage val text: String? = null,
    val server: SourcegraphServerPath? = null,
    val isServerEditable: Boolean = server == null,
    val login: String? = null,
    val isCheckLoginUnique: Boolean = false,
    val token: String? = null,
    val customRequestHeaders: String? = null
)

internal fun CodyLoginRequest.loginWithToken(
    project: Project?,
    parentComponent: Component?
): CodyAuthData? {
  val dialog = SourcegraphTokenLoginDialog(project, parentComponent, isLoginUniqueChecker)
  configure(dialog)

  return dialog.getAuthData()
}

private val CodyLoginRequest.isLoginUniqueChecker: UniqueLoginPredicate
  get() = { login, server ->
    !isCheckLoginUnique ||
        CodyAuthenticationManager.getInstance().isAccountUnique(login, server)
  }

private fun CodyLoginRequest.configure(dialog: BaseLoginDialog) {
  server?.let { dialog.setServer(it.toString(), isServerEditable) }
  login?.let { dialog.setLogin(it) }
  token?.let { dialog.setToken(it) }
  customRequestHeaders?.let { dialog.setCustomRequestHeaders(it) }
}

private fun BaseLoginDialog.getAuthData(): CodyAuthData? {
  DialogManager.show(this)
  return if (isOK)
      CodyAuthData(
          CodyAccount.create(login, server), login, token)
  else null
}
