package com.sourcegraph.common;

import com.intellij.application.options.CodeStyle;
import com.intellij.openapi.application.ReadAction;
import com.intellij.openapi.editor.Editor;
import com.intellij.psi.codeStyle.CommonCodeStyleSettings;
import java.util.Optional;
import org.jetbrains.annotations.NotNull;

public class EditorUtils {

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
}
