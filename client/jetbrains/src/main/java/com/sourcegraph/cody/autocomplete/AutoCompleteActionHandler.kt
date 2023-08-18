package com.sourcegraph.cody.autocomplete

import com.intellij.openapi.actionSystem.DataContext
import com.intellij.openapi.editor.Caret
import com.intellij.openapi.editor.Editor
import com.intellij.openapi.editor.Inlay
import com.intellij.openapi.editor.actionSystem.EditorActionHandler
import com.sourcegraph.cody.agent.CodyAgent
import com.sourcegraph.cody.autocomplete.render.CodyAutoCompleteElementRenderer
import com.sourcegraph.cody.vscode.InlineAutoCompleteItem

open class AutoCompleteActionHandler : EditorActionHandler() {

  override fun isEnabledForCaret(editor: Editor, caret: Caret, dataContext: DataContext?): Boolean {
    // Returns false to fall back to normal action if there is no suggestion at the caret.
    val project = editor.project
    return if (project != null &&
        CodyAutoCompleteManager.isEditorInstanceSupported(editor) &&
        CodyAgent.isConnected(project)) {
      getAgentAutocompleteItem(caret) != null
    } else {
      AutoCompleteText.atCaret(caret).isPresent
    }
  }

  /**
   * Returns the autocompletion item for the first inlay of type `CodyAutoCompleteElementRenderer`
   * regardless if the inlay is positioned at the caret. The reason we don't require the inlay to be
   * positioned at the caret is that completions can suggest changes in a nearby character like in
   * this situation:
   *
   * ` System.out.println("a: CARET"); // original System.out.println("a: " + a);CARET //
   * autocomplete ` *
   */
  protected fun getAgentAutocompleteItem(caret: Caret): InlineAutoCompleteItem? {
    return InlayModelUtils.getAllInlaysForEditor(caret.editor)
        .filter { r: Inlay<*> -> r.renderer is CodyAutoCompleteElementRenderer }
        .firstNotNullOfOrNull { r: Inlay<*> ->
          (r.renderer as CodyAutoCompleteElementRenderer).completionItem
        }
  }
}
