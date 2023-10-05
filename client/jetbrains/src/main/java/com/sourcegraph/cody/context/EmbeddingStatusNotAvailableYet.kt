package com.sourcegraph.cody.context

import com.intellij.openapi.project.Project
import javax.swing.Icon

class EmbeddingStatusNotAvailableYet : EmbeddingStatus {
  override fun getIcon(): Icon? {
    return null
  }

  override fun getTooltip(project: Project): String {
    return ""
  }

  override fun getMainText(): String {
    return ""
  }
}
