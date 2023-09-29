package com.sourcegraph.cody.autocomplete.action

import com.intellij.openapi.actionSystem.DataContext
import com.intellij.openapi.diagnostic.Logger
import com.intellij.openapi.editor.Caret
import com.intellij.openapi.editor.Editor
import com.sourcegraph.utils.CodyEditorUtil

class CycleCodyAutocompleteActionHandler : AutocompleteActionHandler() {
  private val logger = Logger.getInstance(CycleCodyAutocompleteActionHandler::class.java)

  override fun isEnabledForCaret(editor: Editor, caret: Caret, dataContext: DataContext?): Boolean {
    val project = editor.project ?: return false
    return CodyEditorUtil.isEditorInstanceSupported(editor) &&
        hasAnyAutocompleteItems(project, caret)
  }

  override fun doExecute(editor: Editor, caret: Caret?, dataContext: DataContext?) {
    // TODO: implement the actual cycling logic
    logger.warn("cycle trigger")
  }
}
