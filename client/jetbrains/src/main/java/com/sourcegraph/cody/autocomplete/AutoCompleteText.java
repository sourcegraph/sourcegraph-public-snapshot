package com.sourcegraph.cody.autocomplete;

import com.intellij.openapi.editor.Caret;
import com.intellij.openapi.editor.Editor;
import com.intellij.openapi.editor.Inlay;
import com.sourcegraph.cody.autocomplete.render.AutoCompleteRendererType;
import com.sourcegraph.cody.autocomplete.render.CodyAutoCompleteBlockElementRenderer;
import com.sourcegraph.cody.autocomplete.render.CodyAutoCompleteElementRenderer;
import com.sourcegraph.cody.autocomplete.render.CodyAutoCompleteSingleLineRenderer;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;
import org.jetbrains.annotations.NotNull;

public class AutoCompleteText {
  @NotNull public final String sameLineBeforeSuffixText;
  @NotNull public final String sameLineAfterSuffixText;
  @NotNull public final String blockText;

  public AutoCompleteText(
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
   * This method checks for auto completions at a given caret and returns an AutoCompleteTextAtCaret
   * instance if any is found. If any CodyAutoCompleteElementRenderers are present in the inlay
   * model at the given current offset, their corresponding autocompletion will be returned.
   *
   * @param caret the caret at which we look for auto completions
   * @return AutoCompleteText with its corresponding caret if there is any autocompletion to apply
   *     at the caret or Optional.empty otherwise
   */
  public static Optional<AutoCompleteTextAtCaret> atCaret(@NotNull Caret caret) {
    List<CodyAutoCompleteElementRenderer> autoCompleteRenderers =
        InlayModelUtils.getAllInlaysForCaret(caret).stream()
            .map(Inlay::getRenderer)
            .filter(r -> r instanceof CodyAutoCompleteElementRenderer)
            .map(r -> (CodyAutoCompleteElementRenderer) r)
            .collect(Collectors.toList());
    List<CodyAutoCompleteSingleLineRenderer> singleLineRenderers =
        autoCompleteRenderers.stream()
            .filter(r -> r instanceof CodyAutoCompleteSingleLineRenderer)
            .map(r -> (CodyAutoCompleteSingleLineRenderer) r)
            .collect(Collectors.toList());
    List<CodyAutoCompleteBlockElementRenderer> multiLineRenderers =
        autoCompleteRenderers.stream()
            .filter(r -> r instanceof CodyAutoCompleteBlockElementRenderer)
            .map(r -> (CodyAutoCompleteBlockElementRenderer) r)
            .collect(Collectors.toList());

    String inlineText =
        singleLineRenderers.stream()
            .filter(r -> r.getType() == AutoCompleteRendererType.INLINE)
            .findFirst() // note that we only care about the first inline
            .map(CodyAutoCompleteElementRenderer::getText)
            .orElse("");
    String afterEndOfLineText =
        singleLineRenderers.stream()
            .filter(r -> r.getType() == AutoCompleteRendererType.AFTER_LINE_END)
            .findFirst() // note that we only care about the first afterEndOfLine
            .map(CodyAutoCompleteElementRenderer::getText)
            .orElse("");
    String blockText =
        multiLineRenderers.stream()
            .findFirst() // note that we only care about the first block
            .map(CodyAutoCompleteElementRenderer::getText)
            .orElse("");
    if (!inlineText.isEmpty() || !afterEndOfLineText.isEmpty() || !blockText.isEmpty())
      // Only return a non-empty autocomplete if there's at least one non-empty string
      // to apply.
      return Optional.of(
          new AutoCompleteTextAtCaret(caret, inlineText, afterEndOfLineText, blockText));
    else return Optional.empty();
  }

  public Optional<CodyAutoCompleteSingleLineRenderer> getInlineRenderer(@NotNull Editor editor) {
    return this.sameLineBeforeSuffixText.isBlank()
        ? Optional.empty()
        : Optional.of(
            new CodyAutoCompleteSingleLineRenderer(
                this.sameLineBeforeSuffixText, editor, AutoCompleteRendererType.INLINE));
  }

  public Optional<CodyAutoCompleteSingleLineRenderer> getAfterLineEndRenderer(
      @NotNull Editor editor) {
    return this.sameLineAfterSuffixText.isBlank()
        ? Optional.empty()
        : Optional.of(
            new CodyAutoCompleteSingleLineRenderer(
                this.sameLineAfterSuffixText, editor, AutoCompleteRendererType.AFTER_LINE_END));
  }

  public Optional<CodyAutoCompleteBlockElementRenderer> getBlockRenderer(@NotNull Editor editor) {
    return this.blockText.isBlank()
        ? Optional.empty()
        : Optional.of(new CodyAutoCompleteBlockElementRenderer(this.blockText, editor));
  }
}
