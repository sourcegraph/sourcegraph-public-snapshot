package com.sourcegraph.cody.autocomplete;

import com.intellij.openapi.editor.Caret;
import com.sourcegraph.cody.vscode.InlineAutoCompleteItem;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

public class AutoCompleteTextAtCaret {
  @NotNull public final AutoCompleteText autoCompleteText;
  @NotNull public final Caret caret;
  @Nullable public InlineAutoCompleteItem completionItem;

  public AutoCompleteTextAtCaret(@NotNull Caret caret, @NotNull AutoCompleteText autoCompleteText) {
    this.caret = caret;
    this.autoCompleteText = autoCompleteText;
  }

  public AutoCompleteTextAtCaret(
      @NotNull Caret caret,
      @NotNull String sameLineBeforeSuffixText,
      @NotNull String sameLineAfterSuffixText,
      @NotNull String blockText,
      @Nullable InlineAutoCompleteItem completionItem) {
    this(caret, new AutoCompleteText(sameLineBeforeSuffixText, sameLineAfterSuffixText, blockText));
    this.completionItem = completionItem;
  }
}
