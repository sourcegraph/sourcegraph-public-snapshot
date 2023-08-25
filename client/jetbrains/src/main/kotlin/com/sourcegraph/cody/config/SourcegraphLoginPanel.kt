package com.sourcegraph.cody.config

import com.intellij.collaboration.async.CompletableFutureUtil.completionOnEdt
import com.intellij.collaboration.async.CompletableFutureUtil.errorOnEdt
import com.intellij.collaboration.async.CompletableFutureUtil.submitIOTask
import com.intellij.openapi.components.service
import com.intellij.openapi.progress.ProgressIndicator
import com.intellij.openapi.progress.ProgressManager
import com.intellij.openapi.ui.ValidationInfo
import com.intellij.ui.AnimatedIcon
import com.intellij.ui.components.fields.ExtendableTextComponent
import com.intellij.ui.components.fields.ExtendableTextField
import com.intellij.ui.components.panels.Wrapper
import com.intellij.ui.layout.LayoutBuilder
import java.util.concurrent.CompletableFuture
import javax.swing.JComponent
import javax.swing.JTextField

internal typealias UniqueLoginPredicate = (login: String, server: SourcegraphServerPath) -> Boolean

internal class SourcegraphLoginPanel(
    executorFactory: SourcegraphApiRequestExecutor.Factory,
    isAccountUnique: UniqueLoginPredicate
) : Wrapper() {

  private val serverTextField = ExtendableTextField(SourcegraphServerPath.DEFAULT_HOST, 0)
  private var tokenAcquisitionError: ValidationInfo? = null

  private lateinit var currentUi: SourcegraphCredentialsUi
  private var tokenUi =
      SourcegraphTokenCredentialsUi(serverTextField, executorFactory, isAccountUnique)
  private var oauthUi = SourcegraphOAuthCredentialsUi(executorFactory, isAccountUnique)

  private val progressIcon = AnimatedIcon.Default()
  private val progressExtension = ExtendableTextComponent.Extension { progressIcon }

  var footer: LayoutBuilder.() -> Unit
    get() = tokenUi.footer
    set(value) {
      tokenUi.footer = value
      oauthUi.footer = value
      applyUi(currentUi)
    }

  init {
    applyUi(tokenUi)
  }

  private fun applyUi(ui: SourcegraphCredentialsUi) {
    currentUi = ui
    setContent(currentUi.getPanel())
    currentUi.getPreferredFocusableComponent()?.requestFocus()
    tokenAcquisitionError = null
  }

  fun getPreferredFocusableComponent(): JComponent? =
      serverTextField.takeIf { it.isEditable && it.text.isBlank() }
          ?: currentUi.getPreferredFocusableComponent()

  fun doValidateAll(): List<ValidationInfo> {
    val uiError =
        DialogValidationUtils.notBlank(
            serverTextField, "Server url cannot be empty")
            ?: validateServerPath(serverTextField) ?: currentUi.getValidator().invoke()

    return listOfNotNull(uiError, tokenAcquisitionError)
  }

  private fun validateServerPath(field: JTextField): ValidationInfo? =
      try {
        SourcegraphServerPath.from(field.text)
        null
      } catch (e: Exception) {
        ValidationInfo("Invalid server url", field)
      }

  private fun setBusy(busy: Boolean) {
    serverTextField.apply {
      if (busy) addExtension(progressExtension) else removeExtension(progressExtension)
    }
    serverTextField.isEnabled = !busy

    currentUi.setBusy(busy)
  }

  fun acquireLoginAndToken(progressIndicator: ProgressIndicator): CompletableFuture<Pair<String, String>> {
    setBusy(true)
    tokenAcquisitionError = null

    val server = getServer()
    val executor = currentUi.createExecutor()

    return service<ProgressManager>()
        .submitIOTask(progressIndicator) { currentUi.acquireLoginAndToken(server, executor, it) }
        .completionOnEdt(progressIndicator.modalityState) { setBusy(false) }
        .errorOnEdt(progressIndicator.modalityState) { setError(it) }
  }

  fun getServer(): SourcegraphServerPath = SourcegraphServerPath.from(serverTextField.text.trim())

  fun setServer(path: String, editable: Boolean) {
    serverTextField.text = path
    serverTextField.isEditable = editable
  }

  fun setLogin(login: String?, editable: Boolean) {
    tokenUi.setFixedLogin(if (editable) null else login)
  }

  fun setToken(token: String?) = tokenUi.setToken(token.orEmpty())

  fun setError(exception: Throwable?) {
    tokenAcquisitionError = exception?.let { currentUi.handleAcquireError(it) }
  }

  fun setOAuthUi() = applyUi(oauthUi)
  fun setTokenUi() = applyUi(tokenUi)
}
