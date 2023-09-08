package com.sourcegraph.utils

import com.intellij.openapi.project.Project

object ProjectUtil {
    @JvmStatic
    fun isProjectAvailable(project: Project?): Boolean {
        return project != null && !project.isDisposed
    }
}
