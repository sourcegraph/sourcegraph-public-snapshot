package com.sourcegraph.utils

import com.intellij.ide.lightEdit.LightEdit
import com.intellij.openapi.project.Project

object CodyProjectUtil {
    @JvmStatic
    fun isProjectAvailable(project: Project?): Boolean {
        return project != null && !project.isDisposed
    }

    fun isProjectSupported(project: Project?): Boolean {
        return if (isProjectorEnabled()) {
            true
        } else {
            // Light edit is a mode when users can edit a single file with IntelliJ without loading an
            // entire project. We lean on the conservative side for now and don't support Cody for
            // LightEdit projects.
            !LightEdit.owns(project)
        }
    }

    /** Projector is a JetBrains project that runs IntelliJ on a server as a remote IDE. Users
     * interface with IntelliJ through a JavaScript client, either in the browser or an Electron.js
     * client.
     */
    private fun isProjectorEnabled(): Boolean {
        return "true" == System.getProperty("org.jetbrains.projector.server.enable")
    }
}
