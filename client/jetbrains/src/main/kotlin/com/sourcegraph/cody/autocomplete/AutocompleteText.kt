package com.sourcegraph.cody.autocomplete

import com.intellij.openapi.editor.Caret
import com.sourcegraph.cody.autocomplete.render.AutocompleteRendererType
import com.sourcegraph.cody.autocomplete.render.CodyAutocompleteBlockElementRenderer
import com.sourcegraph.cody.autocomplete.render.CodyAutocompleteElementRenderer
import com.sourcegraph.cody.autocomplete.render.CodyAutocompleteSingleLineRenderer
import com.sourcegraph.cody.autocomplete.render.InlayModelUtil

class AutocompleteText(
    private val sameLineBeforeSuffixText: String,
    private val sameLineAfterSuffixText: String,
    private val blockText: String
) {
  fun getAutoCompletionString(sameLineSuffix: String): String {
    val textBelow = if (blockText.isBlank()) "" else System.lineSeparator() + blockText
    return (sameLineBeforeSuffixText + sameLineSuffix + sameLineAfterSuffixText + textBelow)
  }

  companion object {
    /**
     * This method checks for auto completions at a given caret and returns an
     * AutocompleteTextAtCaret instance if any is found. If any CodyAutocompleteElementRenderers are
     * present in the inlay model at the given current offset, their corresponding autocompletion
     * will be returned.
     *
     * @param caret the caret at which we look for auto completions
     * @return AutocompleteText with its corresponding caret if there is any autocompletion to apply
     *   at the caret or null otherwise
     */
    fun atCaret(caret: Caret): AutocompleteTextAtCaret? {
      val autoCompleteRenderers =
          InlayModelUtil.getAllInlaysForCaret(caret)
              .map { it.renderer }
              .filterIsInstance<CodyAutocompleteElementRenderer>()
      val singleLineRenderers =
          autoCompleteRenderers.filterIsInstance<CodyAutocompleteSingleLineRenderer>()
      val multiLineRenderers =
          autoCompleteRenderers.filterIsInstance<CodyAutocompleteBlockElementRenderer>()
      val inlineText =
          singleLineRenderers
              // note that we only care about the first inline
              .firstOrNull { it.type == AutocompleteRendererType.INLINE }
              ?.text
              ?: ""
      val afterEndOfLineText =
          singleLineRenderers
              // note that we only care about the first afterEndOfLine
              .firstOrNull { it.type == AutocompleteRendererType.AFTER_LINE_END }
              ?.text
              ?: ""
      val blockText =
          multiLineRenderers
              // note that we only care about the first block
              .firstOrNull()
              ?.text
              ?: ""
      // Only return a non-empty autocomplete if there's at least one non-empty string to apply.
      return if (inlineText.isNotEmpty() ||
          afterEndOfLineText.isNotEmpty() ||
          blockText.isNotEmpty())
          AutocompleteTextAtCaret(caret, inlineText, afterEndOfLineText, blockText)
      else null
    }
  }
}
