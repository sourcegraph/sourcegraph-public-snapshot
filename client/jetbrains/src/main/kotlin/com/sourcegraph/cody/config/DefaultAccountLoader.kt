package com.sourcegraph.cody.config

import com.intellij.openapi.components.service
import com.intellij.openapi.project.Project

object DefaultAccountLoader {

  @JvmStatic
  fun loadDefaultAccount(project: Project): SourcegraphAccount? {
    val projectDefaultAccountHolder = project.service<SourcegraphProjectDefaultAccountHolder>()
    return projectDefaultAccountHolder.account
  }
}
