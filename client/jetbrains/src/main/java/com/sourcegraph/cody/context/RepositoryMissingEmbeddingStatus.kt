package com.sourcegraph.cody.context

import com.intellij.openapi.project.Project
import com.sourcegraph.cody.Icons
import javax.swing.Icon

class RepositoryMissingEmbeddingStatus(val repoName: String) :
    RepoAvailableEmbeddingStatus(repoName) {

  override fun getIcon(): Icon? {
    return Icons.Repository.NoEmbedding
  }

  override fun getTooltip(project: Project): String {
    return "Repository $repoName is not indexed and has no embeddings"
  }
}
