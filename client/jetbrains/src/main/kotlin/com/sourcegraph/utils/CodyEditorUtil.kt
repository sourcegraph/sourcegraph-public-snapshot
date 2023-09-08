package com.sourcegraph.utils

import com.intellij.application.options.CodeStyle
import com.intellij.openapi.application.ReadAction
import com.intellij.openapi.editor.Document
import com.intellij.openapi.editor.Editor
import com.intellij.openapi.fileEditor.FileEditor
import com.intellij.openapi.fileEditor.FileEditorManager
import com.intellij.openapi.fileEditor.TextEditor
import com.intellij.openapi.project.Project
import com.intellij.openapi.project.ProjectManager
import com.intellij.openapi.util.TextRange
import com.intellij.psi.codeStyle.CommonCodeStyleSettings
import com.intellij.psi.codeStyle.CommonCodeStyleSettings.IndentOptions
import com.sourcegraph.cody.vscode.Range
import java.util.*
import java.util.stream.Collectors
import kotlin.math.min

object CodyEditorUtil {
    const val VIM_EXIT_INSERT_MODE_ACTION = "VimInsertExitModeAction"

    /**
     * Returns a new String, using the given indentation settings to determine the tab size.
     *
     * @param inputText text with tabs to convert to spaces
     * @param indentOptions relevant code style settings
     * @return a new String with all '\t' characters replaced with spaces according to the configured
     * tab size
     */
    @JvmStatic
    fun tabsToSpaces(
        inputText: String, indentOptions: IndentOptions): String {
        val tabReplacement = " ".repeat(indentOptions.TAB_SIZE)
        return inputText.replace("\t".toRegex(), tabReplacement)
    }

    /**
     * @param editor given editor
     * @return Indent options for the given editor, if null falls back to DEFAULT_INDENT_OPTIONS
     */
    @JvmStatic
    fun indentOptions(
        editor: Editor): IndentOptions {
        return Optional.ofNullable(codeStyleSettings(editor).indentOptions)
            .orElse(IndentOptions.DEFAULT_INDENT_OPTIONS)
    }

    /**
     * @param editor given editor
     * @return code style settings for the given editor, if null defaults to default app code style
     * settings
     */
    @JvmStatic
    fun codeStyleSettings(editor: Editor): CommonCodeStyleSettings {
        return ReadAction.compute<CommonCodeStyleSettings, RuntimeException> {
            Optional.ofNullable(CodeStyle.getLanguageSettings(editor))
                .orElse(CodeStyle.getDefaultSettings())
        }
    }

    @JvmStatic
    fun getTextRange(document: Document, range: Range): TextRange {
        val start = min(
            document.getLineEndOffset(range.start.line),
            document.getLineStartOffset(range.start.line) + range.start.character)
        val end = min(
            document.getLineEndOffset(range.end.line),
            document.getLineStartOffset(range.end.line) + range.end.character)
        return TextRange.create(start, end)
    }

    @JvmStatic
    fun getAllOpenEditors(): Set<Editor> {
        return Arrays.stream(ProjectManager.getInstance().openProjects)
            .flatMap { project: Project? -> Arrays.stream(FileEditorManager.getInstance(project!!).allEditors) }
            .filter { fileEditor: FileEditor? -> fileEditor is TextEditor }
            .map { fileEditor: FileEditor -> (fileEditor as TextEditor).editor }
            .collect(Collectors.toSet())
    }
}
