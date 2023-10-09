package com.sourcegraph.cody.agent

import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.project.Project
import com.intellij.openapi.vfs.VirtualFile
import com.sourcegraph.cody.CodyToolWindowContent
import com.sourcegraph.cody.config.CodyProjectSettings
import com.sourcegraph.config.ConfigUtil
import com.sourcegraph.vcs.RepoUtil

class CodyAgentCodebase(
  private val underlying: CodyAgentServer,
  private val project: Project
) {

  // TODO: Support list of repository names instead of just one.
  private val application = ApplicationManager.getApplication()
  private val settings = CodyProjectSettings.getInstance(project)

  init {
    application.executeOnPooledThread {
      onRepositoryName(settings.remoteUrl ?: RepoUtil.findRepositoryName(project, null))
    }
  }

  fun setUrl(url: String) {
    settings.remoteUrl = url
    onPropagateConfiguration()
  }

  fun getUrl(): String? =
    settings.remoteUrl

  fun onFileOpened(project: Project, file: VirtualFile) {
    application.executeOnPooledThread {
      onRepositoryName(settings.remoteUrl ?: RepoUtil.findRepositoryName(project, file))
    }
  }

  private fun onRepositoryName(url: String?) {
    application.invokeLater {
      if (url != null) {
        settings.remoteUrl = url
        onPropagateConfiguration()
      }
    }
  }

  private fun onPropagateConfiguration() {
    CodyToolWindowContent.getInstance(project).embeddingStatusView.updateEmbeddingStatus()
    underlying.configurationDidChange(ConfigUtil.getAgentConfiguration(project))
  }

}
