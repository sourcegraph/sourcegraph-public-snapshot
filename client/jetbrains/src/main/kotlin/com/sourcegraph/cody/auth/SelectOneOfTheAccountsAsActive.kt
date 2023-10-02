package com.sourcegraph.cody.auth

import com.intellij.openapi.project.Project
import com.sourcegraph.cody.CodyToolWindowContent
import com.sourcegraph.cody.config.CodyAuthenticationManager
import com.sourcegraph.cody.config.getFirstAccountOrNull
import com.sourcegraph.cody.initialization.Activity

class SelectOneOfTheAccountsAsActive : Activity {
  override fun runActivity(project: Project) {
    val codyAuthenticationManager = CodyAuthenticationManager.instance
    if (codyAuthenticationManager.getActiveAccount(project) == null) {
      val newActiveAccount = codyAuthenticationManager.getAccounts().getFirstAccountOrNull()
      codyAuthenticationManager.setActiveAccount(project, newActiveAccount)
      val codyToolWindowContent = CodyToolWindowContent.getInstance(project)
      codyToolWindowContent.refreshPanelsVisibility()
    }
  }
}
