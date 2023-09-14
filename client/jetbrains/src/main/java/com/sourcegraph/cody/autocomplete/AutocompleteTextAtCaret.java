package com.sourcegraph.cody.autocomplete;

import com.intellij.openapi.editor.Caret;
import org.jetbrains.annotations.NotNull;

public class AutocompleteTextAtCaret {
  @NotNull public final AutocompleteText autoCompleteText;
  @NotNull public final Caret caret;

  public AutocompleteTextAtCaret(@NotNull Caret caret, @NotNull AutocompleteText autoCompleteText) {
    this.caret = caret;
    this.autoCompleteText = autoCompleteText;
  }

  public AutocompleteTextAtCaret(
      @NotNull Caret caret,
      @NotNull String sameLineBeforeSuffixText,
      @NotNull String sameLineAfterSuffixText,
      @NotNull String blockText) {
    this(caret, new AutocompleteText(sameLineBeforeSuffixText, sameLineAfterSuffixText, blockText));
  }
}
