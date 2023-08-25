package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.credentials.Credentials
import com.intellij.collaboration.auth.services.OAuthCredentialsAcquirer
import com.intellij.collaboration.auth.services.OAuthRequest
import com.intellij.collaboration.auth.services.OAuthServiceBase
import com.intellij.collaboration.auth.services.PkceUtils
import com.intellij.openapi.components.Service
import com.intellij.openapi.components.service
import com.intellij.util.Url
import com.intellij.util.Urls
import java.util.Base64
import java.util.concurrent.CompletableFuture
import org.jetbrains.ide.BuiltInServerManager
import org.jetbrains.ide.RestService

@Service
internal class SourcegraphOAuthService : OAuthServiceBase<Credentials>() {
  override val name: String
    get() = SERVICE_NAME

  fun authorize(): CompletableFuture<Credentials> {
    return authorize(SourcegraphOAuthRequest())
  }

  override fun revokeToken(token: String) {
    TODO("Not yet implemented")
  }

  private class SourcegraphOAuthRequest : OAuthRequest<Credentials> {
    private val port: Int
      get() = BuiltInServerManager.getInstance().port

    private val codeVerifier = PkceUtils.generateCodeVerifier()

    private val codeChallenge =
        PkceUtils.generateShaCodeChallenge(codeVerifier, Base64.getEncoder())

    override val authorizationCodeUrl: Url
      get() =
          Urls.newFromEncoded(
              "http://127.0.0.1:$port/${RestService.PREFIX}/$SERVICE_NAME/authorization_code")

    override val credentialsAcquirer: OAuthCredentialsAcquirer<Credentials> =
        SourcegraphOAuthCredentialsAcquirer(codeVerifier)

    override val authUrlWithParameters: Url =
        AUTHORIZE_URL.addParameters(
            mapOf(
                "code_challenge" to codeChallenge,
                "callback_url" to authorizationCodeUrl.toExternalForm()))

    companion object {
      private val AUTHORIZE_URL: Url
        get() = SERVICE_URL.resolve("authorize")
    }
  }

  companion object {
    private const val SERVICE_NAME = "sourcegraph/oauth"

    val instance: SourcegraphOAuthService
      get() = service()

    val SERVICE_URL: Url =
        Urls.newFromEncoded("https://sourcegraph.com/user/settings/tokens/new/callback")
  }
}
