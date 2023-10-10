package com.sourcegraph.cody.autocomplete

import com.intellij.openapi.editor.Caret

class AutocompleteTextAtCaret(val caret: Caret, val autoCompleteText: AutocompleteText) {
  constructor(
      caret: Caret,
      sameLineBeforeSuffixText: String,
      sameLineAfterSuffixText: String,
      blockText: String
  ) : this(caret, AutocompleteText(sameLineBeforeSuffixText, sameLineAfterSuffixText, blockText))
}
