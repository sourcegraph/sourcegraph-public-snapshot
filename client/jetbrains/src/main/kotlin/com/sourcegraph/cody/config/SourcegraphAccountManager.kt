package com.sourcegraph.cody.config

import com.intellij.collaboration.auth.AccountManagerBase
import com.intellij.openapi.components.Service
import com.intellij.openapi.components.service
import com.sourcegraph.config.ConfigUtil

@Service
class SourcegraphAccountManager :
    AccountManagerBase<SourcegraphAccount, String>(ConfigUtil.SERVICE_DISPLAY_NAME) {
  override fun accountsRepository() = service<SourcegraphPersisentAccounts>()

  override fun deserializeCredentials(credentials: String): String = credentials

  override fun serializeCredentials(credentials: String): String = credentials
}
