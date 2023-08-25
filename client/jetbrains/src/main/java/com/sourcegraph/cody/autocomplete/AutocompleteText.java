package com.sourcegraph.cody.autocomplete;

import com.intellij.openapi.editor.Caret;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.Inlay;
import com.sourcegraph.cody.autocomplete.render.AutocompleteRendererType;
import com.sourcegraph.cody.autocomplete.render.CodyAutocompleteBlockElementRenderer;
import com.sourcegraph.cody.autocomplete.render.CodyAutocompleteElementRenderer;
import com.sourcegraph.cody.autocomplete.render.CodyAutocompleteSingleLineRenderer;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;
import org.jetbrains.annotations.NotNull;

public class AutocompleteText {
  @NotNull public final String sameLineBeforeSuffixText;
  @NotNull public final String sameLineAfterSuffixText;
  @NotNull public final String blockText;

  @Override
  public String toString() {
    return "AutocompleteText{"
        + "sameLineBeforeSuffixText='"
        + sameLineBeforeSuffixText
        + '\''
        + ", sameLineAfterSuffixText='"
        + sameLineAfterSuffixText
        + '\''
        + ", blockText='"
        + blockText
        + '\''
        + '}';
  }

  public AutocompleteText(
      @NotNull String sameLineBeforeSuffixText,
      @NotNull String sameLineAfterSuffixText,
      @NotNull String blockText) {
    this.sameLineBeforeSuffixText = sameLineBeforeSuffixText;
    this.sameLineAfterSuffixText = sameLineAfterSuffixText;
    this.blockText = blockText;
  }

  public @NotNull String getAutoCompletionString(@NotNull String sameLineSuffix) {
    String textBelow = this.blockText.isBlank() ? "" : (System.lineSeparator() + this.blockText);
    return this.sameLineBeforeSuffixText
        + sameLineSuffix
        + this.sameLineAfterSuffixText
        + textBelow;
  }

  /**
   * This method checks for auto completions at a given caret and returns an AutocompleteTextAtCaret
   * instance if any is found. If any CodyAutocompleteElementRenderers are present in the inlay
   * model at the given current offset, their corresponding autocompletion will be returned.
   *
   * @param caret the caret at which we look for auto completions
   * @return AutocompleteText with its corresponding caret if there is any autocompletion to apply
   *     at the caret or Optional.empty otherwise
   */
  public static Optional<AutocompleteTextAtCaret> atCaret(@NotNull Caret caret) {
    List<CodyAutocompleteElementRenderer> autoCompleteRenderers =
        InlayModelUtils.getAllInlaysForCaret(caret).stream()
            .map(Inlay::getRenderer)
            .filter(r -> r instanceof CodyAutocompleteElementRenderer)
            .map(r -> (CodyAutocompleteElementRenderer) r)
            .collect(Collectors.toList());
    List<CodyAutocompleteSingleLineRenderer> singleLineRenderers =
        autoCompleteRenderers.stream()
            .filter(r -> r instanceof CodyAutocompleteSingleLineRenderer)
            .map(r -> (CodyAutocompleteSingleLineRenderer) r)
            .collect(Collectors.toList());
    List<CodyAutocompleteBlockElementRenderer> multiLineRenderers =
        autoCompleteRenderers.stream()
            .filter(r -> r instanceof CodyAutocompleteBlockElementRenderer)
            .map(r -> (CodyAutocompleteBlockElementRenderer) r)
            .collect(Collectors.toList());

    String inlineText =
        singleLineRenderers.stream()
            .filter(r -> r.getType() == AutocompleteRendererType.INLINE)
            .findFirst() // note that we only care about the first inline
            .map(CodyAutocompleteElementRenderer::getText)
            .orElse("");
    String afterEndOfLineText =
        singleLineRenderers.stream()
            .filter(r -> r.getType() == AutocompleteRendererType.AFTER_LINE_END)
            .findFirst() // note that we only care about the first afterEndOfLine
            .map(CodyAutocompleteElementRenderer::getText)
            .orElse("");
    String blockText =
        multiLineRenderers.stream()
            .findFirst() // note that we only care about the first block
            .map(CodyAutocompleteElementRenderer::getText)
            .orElse("");
    if (!inlineText.isEmpty() || !afterEndOfLineText.isEmpty() || !blockText.isEmpty())
      // Only return a non-empty autocomplete if there's at least one non-empty string
      // to apply.
      return Optional.of(
          new AutocompleteTextAtCaret(caret, inlineText, afterEndOfLineText, blockText));
    else return Optional.empty();
  }

  public Optional<CodyAutocompleteSingleLineRenderer> getInlineRenderer(@NotNull Editor editor) {
    return this.sameLineBeforeSuffixText.isBlank()
        ? Optional.empty()
        : Optional.of(
            new CodyAutocompleteSingleLineRenderer(
                this.sameLineBeforeSuffixText, null, editor, AutocompleteRendererType.INLINE));
  }

  public Optional<CodyAutocompleteSingleLineRenderer> getAfterLineEndRenderer(
      @NotNull Editor editor) {
    return this.sameLineAfterSuffixText.isBlank()
        ? Optional.empty()
        : Optional.of(
            new CodyAutocompleteSingleLineRenderer(
                this.sameLineAfterSuffixText,
                null,
                editor,
                AutocompleteRendererType.AFTER_LINE_END));
  }

  public Optional<CodyAutocompleteBlockElementRenderer> getBlockRenderer(@NotNull Editor editor) {
    return this.blockText.isBlank()
        ? Optional.empty()
        : Optional.of(new CodyAutocompleteBlockElementRenderer(this.blockText, null, editor));
  }
}
