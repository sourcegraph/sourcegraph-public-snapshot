package com.sourcegraph.utils

import com.intellij.lang.Language
import com.intellij.lang.LanguageUtil
import com.intellij.openapi.editor.Document
import com.intellij.openapi.editor.Editor
import com.intellij.openapi.fileEditor.FileDocumentManager
import com.intellij.openapi.project.Project
import com.sourcegraph.config.ConfigUtil

object CodyLanguageUtil {
    @JvmStatic
    fun getLanguage(project: Project, document: Document): Language? {
        return LanguageUtil.getLanguageForPsi(
            project,
            FileDocumentManager.getInstance().getFile(document))
    }
}
