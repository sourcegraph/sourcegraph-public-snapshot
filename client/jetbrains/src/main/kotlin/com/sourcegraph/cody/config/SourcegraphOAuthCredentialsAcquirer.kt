package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.credentials.Credentials
import com.intellij.collaboration.auth.credentials.SimpleCredentials
import com.intellij.collaboration.auth.services.OAuthCredentialsAcquirer
import com.intellij.collaboration.auth.services.OAuthCredentialsAcquirerHttp
import com.intellij.util.Url

internal class SourcegraphOAuthCredentialsAcquirer(private val codeVerifier: String) :
    OAuthCredentialsAcquirer<Credentials> {
  override fun acquireCredentials(
      code: String
  ): OAuthCredentialsAcquirer.AcquireCredentialsResult<Credentials> {
    val tokenUrl =
        ACCESS_TOKEN_URL.addParameters(mapOf("code" to code, "code_verifier" to codeVerifier))

    return OAuthCredentialsAcquirerHttp.requestToken(tokenUrl) { _, headers ->
      SimpleCredentials(headers.firstValue("X-OAuth-Token").get())
    }
  }

  companion object {
    private val ACCESS_TOKEN_URL: Url
      get() = SourcegraphOAuthService.SERVICE_URL
  }
}
