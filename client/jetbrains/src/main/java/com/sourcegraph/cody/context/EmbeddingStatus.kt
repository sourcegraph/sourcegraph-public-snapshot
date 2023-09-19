package com.sourcegraph.cody.context

import com.intellij.openapi.project.Project
import javax.swing.Icon

interface EmbeddingStatus {
  fun getIcon(): Icon?

  fun getTooltip(project: Project): String

  fun getMainText(): String
}
