package com.sourcegraph.cody.autocomplete.action

import com.intellij.openapi.actionSystem.DataContext
import com.intellij.openapi.diagnostic.Logger
import com.intellij.openapi.editor.Caret
import com.intellij.openapi.editor.Editor
import com.intellij.openapi.editor.actionSystem.EditorActionHandler
import com.sourcegraph.cody.autocomplete.CodyAutocompleteManager
import com.sourcegraph.cody.vscode.InlineCompletionTriggerKind
import com.sourcegraph.utils.CodyEditorUtil

class TriggerAutocompleteActionHandler : EditorActionHandler() {
  val logger = Logger.getInstance(TriggerAutocompleteActionHandler::class.java)

  override fun isEnabledForCaret(editor: Editor, caret: Caret, dataContext: DataContext): Boolean =
      CodyEditorUtil.isEditorInstanceSupported(editor)

  override fun doExecute(editor: Editor, caret: Caret?, dataContext: DataContext) {
    val offset = caret?.offset ?: editor.caretModel.currentCaret.offset
    CodyAutocompleteManager.instance.triggerAutocomplete(
        editor, offset, InlineCompletionTriggerKind.INVOKE)
  }
}
