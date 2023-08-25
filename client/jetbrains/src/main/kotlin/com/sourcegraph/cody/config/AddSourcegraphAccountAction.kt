package com.sourcegraph.cody.config

import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.actionSystem.PlatformCoreDataKeys.CONTEXT_COMPONENT
import com.intellij.openapi.project.DumbAwareAction
import com.intellij.openapi.project.Project
import com.intellij.util.ui.JBUI.Panels.simplePanel
import java.awt.Component
import javax.swing.Action
import javax.swing.JComponent

class AddSourcegraphAccountAction : DumbAwareAction() {
  override fun update(e: AnActionEvent) {
    e.presentation.isEnabledAndVisible = e.getData(SourcegraphAccountsHost.KEY) != null
  }

  override fun actionPerformed(e: AnActionEvent) {
    val accountsHost = e.getData(SourcegraphAccountsHost.KEY)!!
    val dialog =
        SourcegraphOAuthLoginDialog(
            e.project, e.getData(CONTEXT_COMPONENT), accountsHost::isAccountUnique)
    dialog.setServer(SourcegraphServerPath.DEFAULT_HOST, false)

    if (dialog.showAndGet()) {
      accountsHost.addAccount(dialog.server, dialog.login, dialog.token)
    }
  }
}

internal class SourcegraphOAuthLoginDialog(
    project: Project?,
    parent: Component?,
    isAccountUnique: UniqueLoginPredicate
) :
    BaseLoginDialog(
        project, parent, SourcegraphApiRequestExecutor.Factory.getInstance(), isAccountUnique) {

  init {
    title = "Login to Sourcegraph"
    loginPanel.setOAuthUi()
    init()
  }

  override fun createActions(): Array<Action> = arrayOf(cancelAction)

  override fun show() {
    doOKAction()
    super.show()
  }

  override fun createCenterPanel(): JComponent =
      simplePanel(loginPanel).withPreferredWidth(200).setPaddingCompensated()
}
