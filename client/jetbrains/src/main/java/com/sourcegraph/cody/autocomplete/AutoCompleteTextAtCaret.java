package com.sourcegraph.cody.autocomplete;

import com.intellij.openapi.editor.Caret;
import org.jetbrains.annotations.NotNull;

class AutoCompleteTextAtCaret {
  @NotNull public final AutoCompleteText autoCompleteText;
  @NotNull public final Caret caret;

  public AutoCompleteTextAtCaret(@NotNull Caret caret, @NotNull AutoCompleteText autoCompleteText) {
    this.caret = caret;
    this.autoCompleteText = autoCompleteText;
  }

  public AutoCompleteTextAtCaret(
      @NotNull Caret caret,
      @NotNull String sameLineBeforeSuffixText,
      @NotNull String sameLineAfterSuffixText,
      @NotNull String blockText) {
    this(caret, new AutoCompleteText(sameLineBeforeSuffixText, sameLineAfterSuffixText, blockText));
  }
}
