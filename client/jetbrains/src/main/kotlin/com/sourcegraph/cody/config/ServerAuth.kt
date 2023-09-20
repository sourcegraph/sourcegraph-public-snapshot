package com.sourcegraph.cody.config

import com.intellij.openapi.project.Project
import com.sourcegraph.config.ConfigUtil

data class ServerAuth(
    val instanceUrl: String,
    val accessToken: String,
    val customRequestHeaders: String
)

object ServerAuthLoader {

  @JvmStatic
  fun loadServerAuth(project: Project): ServerAuth {
    val codyAuthenticationManager = CodyAuthenticationManager.getInstance()
    val defaultAccount = codyAuthenticationManager.getActiveAccount(project)
    if (defaultAccount != null) {
      val accessToken = codyAuthenticationManager.getTokenForAccount(defaultAccount) ?: ""
      return ServerAuth(
          defaultAccount.server.url, accessToken, defaultAccount.server.customRequestHeaders)
    }
    return ServerAuth(ConfigUtil.DOTCOM_URL, "", "")
  }
}
