package com.sourcegraph.cody.initialization

import com.intellij.openapi.project.Project

interface Activity {
  fun runActivity(project: Project)
}
