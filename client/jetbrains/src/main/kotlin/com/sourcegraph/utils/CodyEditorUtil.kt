package com.sourcegraph.utils

import com.intellij.application.options.CodeStyle
import com.intellij.injected.editor.EditorWindow
import com.intellij.lang.Language
import com.intellij.openapi.application.ReadAction
import com.intellij.openapi.editor.Document
import com.intellij.openapi.editor.Editor
import com.intellij.openapi.editor.ex.EditorEx
import com.intellij.openapi.editor.impl.ImaginaryEditor
import com.intellij.openapi.fileEditor.FileEditor
import com.intellij.openapi.fileEditor.FileEditorManager
import com.intellij.openapi.fileEditor.TextEditor
import com.intellij.openapi.project.Project
import com.intellij.openapi.project.ProjectManager
import com.intellij.openapi.util.Key
import com.intellij.openapi.util.TextRange
import com.intellij.psi.codeStyle.CommonCodeStyleSettings
import com.intellij.psi.codeStyle.CommonCodeStyleSettings.IndentOptions
import com.intellij.util.concurrency.annotations.RequiresEdt
import com.sourcegraph.cody.vscode.Range
import com.sourcegraph.config.ConfigUtil
import java.util.*
import java.util.stream.Collectors
import kotlin.math.min

object CodyEditorUtil {
  const val VIM_EXIT_INSERT_MODE_ACTION = "VimInsertExitModeAction"

  private const val VIM_MOTION_COMMAND = "Motion"
  private const val UP_COMMAND = "Up"
  private const val DOWN_COMMAND = "Down"
  private const val LEFT_COMMAND = "Left"
  private const val RIGHT_COMMAND = "Right"
  private const val MOVE_CARET_COMMAND = "Move Caret"

  @JvmStatic private val KEY_EDITOR_SUPPORTED = Key.create<Boolean>("cody.editorSupported")

  /**
   * Returns a new String, using the given indentation settings to determine the tab size.
   *
   * @param inputText text with tabs to convert to spaces
   * @param indentOptions relevant code style settings
   * @return a new String with all '\t' characters replaced with spaces according to the configured
   *   tab size
   */
  @JvmStatic
  fun tabsToSpaces(inputText: String, indentOptions: IndentOptions): String {
    val tabReplacement = " ".repeat(indentOptions.TAB_SIZE)
    return inputText.replace("\t".toRegex(), tabReplacement)
  }

  /**
   * @param editor given editor
   * @return Indent options for the given editor, if null falls back to DEFAULT_INDENT_OPTIONS
   */
  @JvmStatic
  fun indentOptions(editor: Editor): IndentOptions {
    return Optional.ofNullable(codeStyleSettings(editor).indentOptions)
        .orElse(IndentOptions.DEFAULT_INDENT_OPTIONS)
  }

  /**
   * @param editor given editor
   * @return code style settings for the given editor, if null defaults to default app code style
   *   settings
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
    val start =
        min(
            document.getLineEndOffset(range.start.line),
            document.getLineStartOffset(range.start.line) + range.start.character)
    val end =
        min(
            document.getLineEndOffset(range.end.line),
            document.getLineStartOffset(range.end.line) + range.end.character)
    return TextRange.create(start, end)
  }

  @JvmStatic
  fun getAllOpenEditors(): Set<Editor> {
    return Arrays.stream(ProjectManager.getInstance().openProjects)
        .flatMap { project: Project? ->
          Arrays.stream(FileEditorManager.getInstance(project!!).allEditors)
        }
        .filter { fileEditor: FileEditor? -> fileEditor is TextEditor }
        .map { fileEditor: FileEditor -> (fileEditor as TextEditor).editor }
        .collect(Collectors.toSet())
  }

  @JvmStatic
  fun isEditorInstanceSupported(editor: Editor): Boolean {
    return editor.project != null &&
        !editor.isViewer &&
        !editor.isOneLineMode &&
        editor !is EditorWindow &&
        editor !is ImaginaryEditor &&
        (editor !is EditorEx || !editor.isEmbeddedIntoDialogWrapper)
  }

  @JvmStatic
  private fun isEditorSupported(editor: Editor): Boolean {
    if (editor.isDisposed) {
      return false
    }
    val fromCache = KEY_EDITOR_SUPPORTED[editor]
    if (fromCache != null) {
      return fromCache
    }
    val isSupported =
        isEditorInstanceSupported(editor) && CodyProjectUtil.isProjectSupported(editor.project)
    KEY_EDITOR_SUPPORTED[editor] = isSupported
    return isSupported
  }

  @JvmStatic
  @RequiresEdt
  fun isEditorValidForAutocomplete(editor: Editor?): Boolean {
    return editor != null &&
        editor.document.isWritable &&
        CodyProjectUtil.isProjectAvailable(editor.project) &&
        isEditorSupported(editor)
  }

  @JvmStatic
  fun isImplicitAutocompleteEnabledForEditor(editor: Editor): Boolean {
    return ConfigUtil.isCodyEnabled() &&
        ConfigUtil.isCodyAutocompleteEnabled() &&
        !isLanguageBlacklisted(editor)
  }

  @JvmStatic
  fun getLanguage(editor: Editor): Language? {
    val project = editor.project ?: return null
    return CodyLanguageUtil.getLanguage(project, editor.document)
  }

  @JvmStatic
  fun isLanguageBlacklisted(editor: Editor): Boolean {
    val language = getLanguage(editor) ?: return false
    return ConfigUtil.getBlacklistedAutocompleteLanguageIds().contains(language.id)
  }

  @JvmStatic
  fun isCommandExcluded(command: String?): Boolean {
    return (command.isNullOrEmpty() ||
        command.contains(VIM_MOTION_COMMAND) ||
        command == UP_COMMAND ||
        command == DOWN_COMMAND ||
        command == LEFT_COMMAND ||
        command == RIGHT_COMMAND ||
        command.contains(MOVE_CARET_COMMAND))
  }
}
