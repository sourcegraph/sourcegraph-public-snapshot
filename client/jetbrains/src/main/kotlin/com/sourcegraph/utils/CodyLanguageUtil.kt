package com.sourcegraph.utils

import com.intellij.lang.Language
import com.intellij.lang.LanguageUtil
import com.intellij.openapi.editor.Document
import com.intellij.openapi.editor.Editor
import com.intellij.openapi.fileEditor.FileDocumentManager
import com.intellij.openapi.project.Project
import com.sourcegraph.config.ConfigUtil

class CodyLanguageUtil {
    companion object {
        fun getLanguage(project: Project, document: Document): Language? {
            return LanguageUtil.getLanguageForPsi(
                project,
                FileDocumentManager.getInstance().getFile(document))
        }

        fun getLanguage(editor: Editor): Language? {
            val project = editor.project ?: return null
            return getLanguage(project, editor.document)
        }

        fun isLanguageBlacklisted(editor: Editor): Boolean {
            val language = getLanguage(editor) ?: return false
            return ConfigUtil.getBlacklistedAutocompleteLanguageIds().contains(language.id)
        }
    }
}
