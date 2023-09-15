package com.sourcegraph.cody.agent

import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.project.Project
import com.intellij.openapi.vfs.VirtualFile
import com.sourcegraph.cody.CodyToolWindowContent
import com.sourcegraph.config.ConfigUtil
import com.sourcegraph.vcs.RepoUtil

class CodyAgentCodebase(private val underlying: CodyAgentServer, val project: Project) {

  // TODO: This should ideally be persisted as a project-wide setting. Down the road, it should also
  // support a list of repository names instead of only a single codebase.
  var currentCodebase: String? = null

  init {
    ApplicationManager.getApplication().executeOnPooledThread {
      onRepositoryName(project, RepoUtil.findRepositoryName(project, null))
    }
  }

  fun onFileOpened(project: Project, file: VirtualFile) {
    ApplicationManager.getApplication().executeOnPooledThread {
      onRepositoryName(project, RepoUtil.findRepositoryName(project, file))
    }
  }

  private fun onRepositoryName(project: Project, repositoryName: String?) {
    ApplicationManager.getApplication().invokeLater {
      if (repositoryName != null && currentCodebase != repositoryName) {
        currentCodebase = repositoryName
        CodyToolWindowContent.getInstance(project).embeddingStatusView.updateEmbeddingStatus()
        underlying.configurationDidChange(ConfigUtil.getAgentConfiguration(project))
      }
    }
  }
}
