package com.sourcegraph.common;

import com.intellij.application.options.CodeStyle;
import com.intellij.openapi.application.ReadAction;
import com.intellij.openapi.editor.Document;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.fileEditor.FileEditorManager;
import com.intellij.openapi.fileEditor.TextEditor;
import com.intellij.openapi.project.ProjectManager;
import com.intellij.openapi.util.TextRange;
import com.intellij.psi.codeStyle.CommonCodeStyleSettings;
import com.sourcegraph.cody.vscode.Range;
import java.util.Arrays;
import java.util.Optional;
import java.util.Set;
import java.util.stream.Collectors;
import org.jetbrains.annotations.NotNull;

public class EditorUtils {

  public static final String VIM_EXIT_INSERT_MODE_ACTION = "VimInsertExitModeAction";

  /**
   * Returns a new String, using the given indentation settings to determine the tab size.
   *
   * @param inputText text with tabs to convert to spaces
   * @param indentOptions relevant code style settings
   * @return a new String with all '\t' characters replaced with spaces according to the configured
   *     tab size
   */
  public static @NotNull String tabsToSpaces(
      @NotNull String inputText, @NotNull CommonCodeStyleSettings.IndentOptions indentOptions) {
    String tabReplacement = " ".repeat(indentOptions.TAB_SIZE);
    return inputText.replaceAll("\t", tabReplacement);
  }

  /**
   * @param editor given editor
   * @return Indent options for the given editor, if null falls back to DEFAULT_INDENT_OPTIONS
   */
  public static @NotNull CommonCodeStyleSettings.IndentOptions indentOptions(
      @NotNull Editor editor) {
    return Optional.ofNullable(codeStyleSettings(editor).getIndentOptions())
        .orElse(CommonCodeStyleSettings.IndentOptions.DEFAULT_INDENT_OPTIONS);
  }

  /**
   * @param editor given editor
   * @return code style settings for the given editor, if null defaults to default app code style
   *     settings
   */
  public static @NotNull CommonCodeStyleSettings codeStyleSettings(@NotNull Editor editor) {
    return ReadAction.compute(
        () ->
            Optional.ofNullable(CodeStyle.getLanguageSettings(editor))
                .orElse(CodeStyle.getDefaultSettings()));
  }

  public static @NotNull TextRange getTextRange(Document document, Range range) {
    int start =
        Math.min(
            document.getLineEndOffset(range.start.line),
            document.getLineStartOffset(range.start.line) + range.start.character);
    int end =
        Math.min(
            document.getLineEndOffset(range.end.line),
            document.getLineStartOffset(range.end.line) + range.end.character);
    return TextRange.create(start, end);
  }

  public static @NotNull Set<Editor> getAllOpenEditors() {
    return Arrays.stream(ProjectManager.getInstance().getOpenProjects())
        .flatMap(project -> Arrays.stream(FileEditorManager.getInstance(project).getAllEditors()))
        .filter(fileEditor -> fileEditor instanceof TextEditor)
        .map(fileEditor -> ((TextEditor) fileEditor).getEditor())
        .collect(Collectors.toSet());
  }
}
