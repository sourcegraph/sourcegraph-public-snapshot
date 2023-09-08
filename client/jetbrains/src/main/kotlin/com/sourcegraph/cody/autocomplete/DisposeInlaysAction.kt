package com.sourcegraph.cody.autocomplete

import com.intellij.openapi.actionSystem.DataContext
import com.intellij.openapi.editor.Caret
import com.intellij.openapi.editor.Editor
import com.intellij.openapi.editor.actionSystem.EditorAction
import com.intellij.openapi.project.DumbAware

class DisposeInlaysAction : EditorAction(DisposeInlaysActionHandler()), DumbAware {
  init {
    setInjectedContext(true)
  }
}

class DisposeInlaysActionHandler : AutocompleteActionHandler() {
  override fun doExecute(editor: Editor, caret: Caret?, dataContext: DataContext?) {
    CodyAutocompleteManager.getInstance().disposeInlays(editor)
  }
}
