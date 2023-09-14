package com.sourcegraph.cody.config

import com.intellij.openapi.components.Service
import com.intellij.openapi.components.service
import com.sourcegraph.cody.auth.AccountManagerBase
import com.sourcegraph.config.ConfigUtil

@Service
class CodyAccountManager :
    AccountManagerBase<CodyAccount, String>(ConfigUtil.SERVICE_DISPLAY_NAME) {
  override fun accountsRepository() = service<CodyPersisentAccounts>()

  override fun deserializeCredentials(credentials: String): String = credentials

  override fun serializeCredentials(credentials: String): String = credentials
}
