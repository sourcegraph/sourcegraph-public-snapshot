package com.sourcegraph.cody.autocomplete

import com.intellij.openapi.actionSystem.DataContext
import com.intellij.openapi.editor.Caret
import com.intellij.openapi.editor.Editor

class DisposeInlaysActionHandler : AutocompleteActionHandler() {
  override fun doExecute(editor: Editor, caret: Caret?, dataContext: DataContext?) =
      CodyAutocompleteManager.instance.disposeInlays(editor)
}
