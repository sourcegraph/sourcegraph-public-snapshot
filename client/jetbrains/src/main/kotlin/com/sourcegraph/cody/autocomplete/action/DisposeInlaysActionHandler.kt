package com.sourcegraph.cody.autocomplete.action

import com.intellij.openapi.actionSystem.DataContext
import com.intellij.openapi.editor.Caret
import com.intellij.openapi.editor.Editor
import com.sourcegraph.cody.autocomplete.CodyAutocompleteManager

class DisposeInlaysActionHandler : AutocompleteActionHandler() {
  override fun doExecute(editor: Editor, caret: Caret?, dataContext: DataContext?) =
      CodyAutocompleteManager.instance.disposeInlays(editor)
}
