package com.sourcegraph.cody.agent

import com.intellij.openapi.project.Project
import com.sourcegraph.cody.statusbar.CodyAutocompleteStatusService
import java.util.concurrent.TimeUnit

object CodyAgentManager {
  @JvmStatic
  fun tryRestartingAgentIfNotRunning(project: Project) {
    if (!CodyAgent.isConnected(project)) {
      startAgent(project)
      try {
        CodyAgent.getInitializedServer(project)[3, TimeUnit.SECONDS]
      } catch (ignored: Exception) {
        //
      }
    }
  }

  @JvmStatic
  fun startAgent(project: Project) {
    try {
      if (project.isDisposed) {
        return
      }
      val service = project.getService(CodyAgent::class.java) ?: return
      if (CodyAgent.isConnected(project)) {
        return
      }
      service.initialize()
    } finally {
      CodyAutocompleteStatusService.resetApplication(project)
    }
  }

  @JvmStatic
  fun stopAgent(project: Project) {
    if (project.isDisposed) {
      return
    }
    val service = project.getService(CodyAgent::class.java) ?: return
    service.shutdown()
  }

  @JvmStatic
  fun restartAgent(project: Project) {
    stopAgent(project)
    startAgent(project)
  }
}
