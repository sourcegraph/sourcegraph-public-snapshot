package com.sourcegraph.cody.config

import com.intellij.openapi.components.service
import com.intellij.openapi.project.Project

class AccountsHostProjectHolder {
  var accountsHost: CodyAccountsHost? = null

  companion object {
    @JvmStatic
    fun getInstance(project: Project): AccountsHostProjectHolder {
      return project.service<AccountsHostProjectHolder>()
    }

    @JvmStatic
    fun useAccountsHost(project: Project, block: (CodyAccountsHost) -> Unit) {
      val accountsHostProjectHolder = getInstance(project)
      val host = accountsHostProjectHolder.accountsHost ?: return
      block(host)
      accountsHostProjectHolder.accountsHost = null
    }
  }
}
