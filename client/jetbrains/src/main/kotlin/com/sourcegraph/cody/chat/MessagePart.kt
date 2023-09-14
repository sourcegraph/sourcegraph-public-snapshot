package com.sourcegraph.cody.chat

import com.intellij.ide.highlighter.HighlighterFactory
import com.intellij.lang.Language
import com.intellij.openapi.command.WriteCommandAction
import com.intellij.openapi.editor.colors.EditorColorsManager
import com.intellij.openapi.editor.ex.EditorEx
import com.intellij.openapi.fileTypes.FileType
import com.intellij.openapi.fileTypes.FileTypeManager
import com.intellij.openapi.fileTypes.PlainTextFileType
import com.intellij.openapi.project.Project
import com.intellij.openapi.util.Computable
import com.intellij.util.ui.SwingHelper
import com.intellij.util.ui.UIUtil
import javax.swing.JComponent
import javax.swing.JEditorPane

sealed interface MessagePart

class TextPart(val component: JEditorPane) : MessagePart {
  fun updateText(text: String) {
    SwingHelper.setHtml(component, text, UIUtil.getLabelForeground())
  }
}

class CodeEditorPart(val component: JComponent, private val editor: EditorEx) : MessagePart {

  fun updateCode(project: Project, code: String, language: String?) {
    updateLanguage(language)
    updateText(project, code)
  }

  fun updateLanguage(language: String?) {
    val fileType: FileType =
        Language.getRegisteredLanguages()
            .firstOrNull { it.displayName.equals(language, ignoreCase = true) }
            ?.let { FileTypeManager.getInstance().findFileTypeByLanguage(it) }
            ?: PlainTextFileType.INSTANCE
    val editorHighlighter =
        HighlighterFactory.createHighlighter(
            fileType, EditorColorsManager.getInstance().schemeForCurrentUITheme, null)
    editor.highlighter = editorHighlighter
  }

  private fun updateText(project: Project, text: String) {
    WriteCommandAction.runWriteCommandAction(
        project, Computable { editor.document.replaceText(text, System.currentTimeMillis()) })
  }
}
