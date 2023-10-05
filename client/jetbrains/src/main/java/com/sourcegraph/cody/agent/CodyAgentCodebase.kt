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
  private var inferredCodebase: String = ""
  private var explicitCodebase: String = ""

  init {
    ApplicationManager.getApplication().executeOnPooledThread {
      onRepositoryName(RepoUtil.findRepositoryName(project, null))
    }
  }

  fun onNewExplicitCodebase(codebase: String) {
    explicitCodebase = codebase
    onPropagateConfiguration()
  }

  fun currentCodebase(): String? = explicitCodebase.ifEmpty { inferredCodebase }.ifEmpty { null }

  fun onFileOpened(project: Project, file: VirtualFile) {
    ApplicationManager.getApplication().executeOnPooledThread {
      onRepositoryName(RepoUtil.findRepositoryName(project, file))
    }
  }

  private fun onPropagateConfiguration() {
    CodyToolWindowContent.getInstance(project).embeddingStatusView.updateEmbeddingStatus()
    underlying.configurationDidChange(ConfigUtil.getAgentConfiguration(project))
  }

  private fun onRepositoryName(repositoryName: String?) {
    ApplicationManager.getApplication().invokeLater {
      if (repositoryName != null && inferredCodebase != repositoryName) {
        inferredCodebase = repositoryName
        onPropagateConfiguration()
      }
    }
  }
}
