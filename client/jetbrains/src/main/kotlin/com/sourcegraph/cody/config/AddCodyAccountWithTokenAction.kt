package com.sourcegraph.cody.config

import com.intellij.ide.BrowserUtil
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.actionSystem.PlatformCoreDataKeys
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.project.DumbAwareAction
import com.intellij.openapi.project.Project
import com.sourcegraph.cody.api.SourcegraphApiRequestExecutor
import com.sourcegraph.config.ConfigUtil
import java.awt.Component
import javax.swing.JComponent
import org.jetbrains.builtInWebServer.BuiltInServerOptions

class AddCodyAccountWithTokenAction : BaseAddAccountWithTokenAction() {
  override val defaultServer: String
    get() = SourcegraphServerPath.DEFAULT_HOST

  override fun actionPerformed(e: AnActionEvent) {
    val accountsHost = e.getData(CodyAccountsHost.KEY)!!
    val project = e.project ?: return
    AccountsHostProjectHolder.getInstance(project).accountsHost = accountsHost
    val port =
        ApplicationManager.getApplication()
            .getService(BuiltInServerOptions::class.java)
            .getEffectiveBuiltInServerPort()
    BrowserUtil.browse(
        ConfigUtil.DOTCOM_URL +
            "user/settings/tokens/new/callback" +
            "?requestFrom=JETBRAINS" +
            "&port=" +
            port)
  }
}

class AddCodyEnterpriseAccountAction : BaseAddAccountWithTokenAction() {
  override val defaultServer: String
    get() = ""

  override fun actionPerformed(e: AnActionEvent) {
    val accountsHost = e.getData(CodyAccountsHost.KEY)!!
    val dialog =
        newAddAccountDialog(
            e.project,
            e.getData(PlatformCoreDataKeys.CONTEXT_COMPONENT),
            accountsHost::isAccountUnique)

    dialog.setServer(defaultServer)
    if (dialog.showAndGet()) {
      accountsHost.addAccount(dialog.server, dialog.login, dialog.displayName, dialog.token)
    }
  }
}

abstract class BaseAddAccountWithTokenAction : DumbAwareAction() {
  abstract val defaultServer: String

  override fun update(e: AnActionEvent) {
    e.presentation.isEnabledAndVisible = e.getData(CodyAccountsHost.KEY) != null
  }
}

private fun newAddAccountDialog(
    project: Project?,
    parent: Component?,
    isAccountUnique: UniqueLoginPredicate
): BaseLoginDialog =
    SourcegraphTokenLoginDialog(project, parent, isAccountUnique).apply {
      title = "Add Sourcegraph Account"
      setLoginButtonText("Add Account")
    }

fun signInWithSourcegrapDialog(
    project: Project?,
    parent: Component?,
    isAccountUnique: UniqueLoginPredicate
): BaseLoginDialog =
    SourcegraphTokenLoginDialog(project, parent, isAccountUnique).apply {
      title = "Sign in with Sourcegraph"
      setLoginButtonText("Sign in")
    }

internal class SourcegraphTokenLoginDialog(
    project: Project?,
    parent: Component?,
    isAccountUnique: UniqueLoginPredicate
) :
    BaseLoginDialog(
        project, parent, SourcegraphApiRequestExecutor.Factory.getInstance(), isAccountUnique) {

  init {
    title = "Login to Sourcegraph"
    setLoginButtonText("Login")
    loginPanel.setTokenUi()
    init()
  }

  override fun createCenterPanel(): JComponent = loginPanel
}
