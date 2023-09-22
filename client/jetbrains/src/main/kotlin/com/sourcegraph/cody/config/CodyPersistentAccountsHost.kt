package com.sourcegraph.cody.config

import com.intellij.openapi.project.Project

class CodyPersistentAccountsHost(private val project: Project?) : CodyAccountsHost {
  override fun addAccount(
      server: SourcegraphServerPath,
      login: String,
      displayName: String?,
      token: String
  ) {
    val codyAccount = CodyAccount.create(login, displayName, server)
    CodyAuthenticationManager.getInstance().updateAccountToken(codyAccount, token)
    if (project != null) {
      CodyAuthenticationManager.getInstance().setActiveAccount(project, codyAccount)
    }
  }

  override fun isAccountUnique(login: String, server: SourcegraphServerPath): Boolean {
    return CodyAuthenticationManager.getInstance().isAccountUnique(login, server)
  }
}
