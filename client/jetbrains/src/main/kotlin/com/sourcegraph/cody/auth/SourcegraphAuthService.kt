package com.sourcegraph.cody.auth

import com.intellij.openapi.components.Service
import com.intellij.openapi.components.service
import com.intellij.util.Url
import com.intellij.util.Urls
import com.sourcegraph.config.ConfigUtil
import java.util.concurrent.CompletableFuture
import org.jetbrains.ide.BuiltInServerManager

@Service
internal class SourcegraphAuthService : AuthServiceBase() {
  override val name: String
    get() = SERVICE_NAME

  fun authorize(): CompletableFuture<String> {
    return authorize(SourcegraphAuthRequest(name))
  }

  private class SourcegraphAuthRequest(override val serviceName: String) : AuthRequest {
    private val port: Int
      get() = BuiltInServerManager.getInstance().port

    override val authUrlWithParameters: Url =
        SERVICE_URL.addParameters(mapOf("requestFrom" to "JETBRAINS", "port" to port.toString()))
  }

  companion object {
    private const val SERVICE_NAME = "sourcegraph"

    val instance: SourcegraphAuthService
      get() = service()

    val SERVICE_URL: Url =
        Urls.newFromEncoded(ConfigUtil.DOTCOM_URL + "user/settings/tokens/new/callback")
  }
}
