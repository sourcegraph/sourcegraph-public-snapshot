package com.sourcegraph.cody.context

import com.intellij.openapi.project.Project
import com.sourcegraph.cody.Icons
import javax.swing.Icon

class NoGitRepositoryEmbeddingStatus : EmbeddingStatus {
  override fun getIcon(): Icon? {
    return Icons.Repository.Missing
  }

  override fun getTooltip(project: Project): String {
    return "No Git repository opened"
  }

  override fun getMainText(): String {
    return ""
  }
}
