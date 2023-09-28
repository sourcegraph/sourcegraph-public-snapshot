package com.sourcegraph.cody.auth

import com.intellij.ide.BrowserUtil
import java.util.concurrent.CompletableFuture
import java.util.concurrent.atomic.AtomicReference

/** The basic service that implements general authorization flow methods */
abstract class AuthServiceBase : AuthService {
  private val currentRequest = AtomicReference<OAuthRequestWithResult?>()

  override fun authorize(request: AuthRequest): CompletableFuture<String> {
    if (!currentRequest.compareAndSet(
        null, OAuthRequestWithResult(request, CompletableFuture<String>()))) {
      return currentRequest.get()!!.result
    }

    val result = currentRequest.get()!!.result
    result.whenComplete { _, _ -> currentRequest.set(null) }
    startAuthorization(request)

    return result
  }

  override fun handleServerCallback(path: String, accessToken: String): Boolean {
    val request = currentRequest.get() ?: return false
    request.processToken(accessToken)
    val result = request.result
    return result.isDone && !result.isCancelled && !result.isCompletedExceptionally
  }

  protected open fun startAuthorization(request: AuthRequest) {
    val authUrl = request.authUrlWithParameters.toExternalForm()
    BrowserUtil.browse(authUrl)
  }

  private fun OAuthRequestWithResult.processToken(accessToken: String) {
    result.complete(accessToken)
  }

  protected data class OAuthRequestWithResult(
      val request: AuthRequest,
      val result: CompletableFuture<String>
  )
}
