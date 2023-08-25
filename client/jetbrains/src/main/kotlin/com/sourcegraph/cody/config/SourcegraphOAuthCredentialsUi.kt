package com.sourcegraph.cody.config

import com.intellij.openapi.progress.ProcessCanceledException
import com.intellij.openapi.progress.ProgressIndicator
import com.intellij.openapi.progress.util.ProgressIndicatorUtils
import com.intellij.openapi.ui.ValidationInfo
import com.intellij.ui.AnimatedIcon
import com.intellij.ui.components.JBLabel
import com.intellij.ui.layout.LayoutBuilder
import com.intellij.util.ui.UIUtil
import javax.swing.JComponent

internal class SourcegraphOAuthCredentialsUi(
    val factory: SourcegraphApiRequestExecutor.Factory,
    val isAccountUnique: UniqueLoginPredicate
) : SourcegraphCredentialsUi() {

  override fun getPreferredFocusableComponent(): JComponent? = null

  override fun getValidator(): Validator = { null }

  override fun createExecutor(): SourcegraphApiRequestExecutor = factory.create("")

  override fun acquireLoginAndToken(
      server: SourcegraphServerPath,
      executor: SourcegraphApiRequestExecutor,
      indicator: ProgressIndicator
  ): Pair<String, String> {
    executor as SourcegraphApiRequestExecutor.WithTokenAuth
    val token = acquireToken(indicator)
    executor.token = token
    val login =
        SourcegraphTokenCredentialsUi.acquireLogin(
            server, executor, indicator, isAccountUnique, null)

    return login to token
  }

  override fun handleAcquireError(error: Throwable): ValidationInfo =
      SourcegraphTokenCredentialsUi.handleError(error)

  override fun setBusy(busy: Boolean) = Unit

  override fun LayoutBuilder.centerPanel() {
    row {
      val progressLabel =
          JBLabel("Logging in\u2026").apply {
            icon = AnimatedIcon.Default()
            foreground = UIUtil.getInactiveTextColor()
          }
      progressLabel()
    }
  }

  private fun acquireToken(indicator: ProgressIndicator): String {
    val credentialsFuture = SourcegraphOAuthService.instance.authorize()
    try {
      return ProgressIndicatorUtils.awaitWithCheckCanceled(credentialsFuture, indicator).accessToken
    } catch (pce: ProcessCanceledException) {
      credentialsFuture.completeExceptionally(pce)
      throw pce
    }
  }
}
