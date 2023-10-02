package com.sourcegraph.cody.config

import com.intellij.collaboration.async.CompletableFutureUtil
import com.intellij.collaboration.async.CompletableFutureUtil.completionOnEdt
import com.intellij.collaboration.async.CompletableFutureUtil.errorOnEdt
import com.intellij.collaboration.async.CompletableFutureUtil.successOnEdt
import com.intellij.openapi.application.ModalityState
import com.intellij.openapi.progress.EmptyProgressIndicator
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.DialogWrapper
import com.intellij.openapi.ui.ValidationInfo
import com.intellij.openapi.util.Disposer
import com.sourcegraph.cody.api.SourcegraphApiRequestExecutor
import java.awt.Component
import javax.swing.JComponent

abstract class BaseLoginDialog(
    project: Project?,
    parent: Component?,
    executorFactory: SourcegraphApiRequestExecutor.Factory,
    isAccountUnique: UniqueLoginPredicate
) : DialogWrapper(project, parent, false, IdeModalityType.PROJECT) {

  protected val loginPanel = CodyLoginPanel(executorFactory, isAccountUnique)

  var login: String = ""
    private set

  var displayName: String? = null
    private set

  var token: String = ""
    private set

  val server: SourcegraphServerPath
    get() = loginPanel.getServer()

  fun setToken(token: String?) = loginPanel.setToken(token)

  fun setLogin(login: String?) = loginPanel.setLogin(login, false)

  fun setServer(path: String) = loginPanel.setServer(path)

  fun setCustomRequestHeaders(customRequestHeaders: String) =
      loginPanel.setCustomRequestHeaders(customRequestHeaders)

  fun setLoginButtonText(text: String) = setOKButtonText(text)

  fun setError(exception: Throwable) {
    loginPanel.setError(exception)
    startTrackingValidation()
  }

  override fun getPreferredFocusedComponent(): JComponent? =
      loginPanel.getPreferredFocusableComponent()

  override fun doValidateAll(): List<ValidationInfo> = loginPanel.doValidateAll()

  override fun doOKAction() {
    val modalityState = ModalityState.stateForComponent(loginPanel)
    val emptyProgressIndicator = EmptyProgressIndicator(modalityState)
    Disposer.register(disposable) { emptyProgressIndicator.cancel() }

    startGettingToken()
    loginPanel
        .acquireDetailsAndToken(emptyProgressIndicator)
        .completionOnEdt(modalityState) { finishGettingToken() }
        .successOnEdt(modalityState) { (details, newToken) ->
          login = details.username
          displayName = details.displayName
          token = newToken

          close(OK_EXIT_CODE, true)
        }
        .errorOnEdt(modalityState) {
          if (!CompletableFutureUtil.isCancellation(it)) startTrackingValidation()
        }
  }

  protected open fun startGettingToken() = Unit

  protected open fun finishGettingToken() = Unit
}
