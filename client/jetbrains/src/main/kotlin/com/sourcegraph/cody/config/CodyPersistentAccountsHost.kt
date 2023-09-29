package com.sourcegraph.cody.config

import com.intellij.openapi.project.Project
import com.sourcegraph.cody.CodyToolWindowContent
import com.sourcegraph.cody.agent.CodyAgent
import com.sourcegraph.config.ConfigUtil

class CodyPersistentAccountsHost(private val project: Project?) : CodyAccountsHost {
  override fun addAccount(
      server: SourcegraphServerPath,
      login: String,
      displayName: String?,
      token: String
  ) {
    val codyAccount = CodyAccount.create(login, displayName, server)
    CodyAuthenticationManager.instance.updateAccountToken(codyAccount, token)
    if (project != null) {
      CodyAuthenticationManager.instance.setActiveAccount(project, codyAccount)
      // Notify Cody Agent about config changes.
      CodyAgent.getServer(project)
          ?.configurationDidChange(ConfigUtil.getAgentConfiguration(project))
      val codyToolWindowContent = CodyToolWindowContent.getInstance(project)
      codyToolWindowContent.refreshPanelsVisibility()
      codyToolWindowContent.embeddingStatusView.updateEmbeddingStatus()
    }
  }

  override fun isAccountUnique(login: String, server: SourcegraphServerPath): Boolean {
    return CodyAuthenticationManager.instance.isAccountUnique(login, server)
  }
}
