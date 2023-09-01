package com.sourcegraph.cody.config

import com.intellij.openapi.components.service
import com.intellij.openapi.project.Project
import com.sourcegraph.config.ConfigUtil

data class ServerAuth(val instanceUrl: String, val accessToken: String, val customRequestHeaders: String)

object ServerAuthLoader {

  @JvmStatic
  fun loadServerAuth(project: Project): ServerAuth {
    val defaultAccountHolder = project.service<SourcegraphProjectDefaultAccountHolder>()
    val sourcegraphAccountManager = service<SourcegraphAccountManager>()
    val defaultAccount = defaultAccountHolder.account
    if (defaultAccount != null) {
      val accessToken = sourcegraphAccountManager.findCredentials(defaultAccount) ?: ""
      return ServerAuth(defaultAccount.server.url, accessToken, defaultAccount.server.customRequestHeaders)
    }
    return ServerAuth(ConfigUtil.DOTCOM_URL, "", "")
  }
}
