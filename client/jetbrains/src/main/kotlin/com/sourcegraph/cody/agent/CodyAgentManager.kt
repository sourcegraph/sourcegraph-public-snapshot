package com.sourcegraph.cody.agent

import com.intellij.openapi.project.Project

object CodyAgentManager {
  @JvmStatic
  fun startAgent(project: Project) {
    if (project.isDisposed) {
      return
    }
    val service = project.getService(CodyAgent::class.java) ?: return
    if (CodyAgent.isConnected(project)) {
      return
    }
    service.initialize()
  }

  @JvmStatic
  fun stopAgent(project: Project) {
    if (project.isDisposed) {
      return
    }
    val service = project.getService(CodyAgent::class.java) ?: return
    service.shutdown()
  }
}
