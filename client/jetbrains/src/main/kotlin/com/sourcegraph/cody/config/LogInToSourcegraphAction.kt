package com.sourcegraph.cody.config

import com.intellij.ide.DataManager
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.actionSystem.PlatformCoreDataKeys
import com.intellij.openapi.project.DumbAwareAction
import com.intellij.openapi.project.Project
import com.intellij.util.ui.JBUI
import com.sourcegraph.cody.api.SourcegraphApiRequestExecutor
import java.awt.Component
import javax.swing.Action
import javax.swing.JComponent

class LogInToSourcegraphAction : BaseAddAccountWithTokenAction() {
  override val defaultServer: String
    get() = SourcegraphServerPath.DEFAULT_HOST

  override fun actionPerformed(e: AnActionEvent) {
    val accountsHost = getCodyAccountsHost(e) ?: return
    val dialog =
        CodyAuthLoginDialog(
            e.project,
            e.getData(PlatformCoreDataKeys.CONTEXT_COMPONENT),
            accountsHost::isAccountUnique)
    dialog.setServer(defaultServer)
    if (dialog.showAndGet()) {
      accountsHost.addAccount(dialog.server, dialog.login, dialog.displayName, dialog.token)
    }
  }
}

class AddCodyEnterpriseAccountAction : BaseAddAccountWithTokenAction() {
  override val defaultServer: String
    get() = ""

  override fun actionPerformed(e: AnActionEvent) {
    val accountsHost = getCodyAccountsHost(e) ?: return
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
    val codyAccountsHost = getCodyAccountsHost(e)
    e.presentation.isEnabledAndVisible = codyAccountsHost != null
  }

  protected fun getCodyAccountsHost(e: AnActionEvent) =
      (e.getData(CodyAccountsHost.DATA_KEY)
          ?: DataManager.getInstance().loadFromDataContext(e.dataContext, CodyAccountsHost.KEY))
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
        project, parent, SourcegraphApiRequestExecutor.Factory.instance, isAccountUnique) {

  init {
    title = "Login to Sourcegraph"
    setLoginButtonText("Login")
    loginPanel.setTokenUi()
    init()
  }

  override fun createCenterPanel(): JComponent = loginPanel
}

internal class CodyAuthLoginDialog(
    project: Project?,
    parent: Component?,
    isAccountUnique: UniqueLoginPredicate
) :
    BaseLoginDialog(
        project, parent, SourcegraphApiRequestExecutor.Factory.instance, isAccountUnique) {

  init {
    title = "Login to Sourcegraph"
    loginPanel.setAuthUI()
    init()
  }

  override fun createActions(): Array<Action> = arrayOf(cancelAction)

  override fun show() {
    doOKAction()
    super.show()
  }

  override fun createCenterPanel(): JComponent =
      JBUI.Panels.simplePanel(loginPanel).withPreferredWidth(200)
}
