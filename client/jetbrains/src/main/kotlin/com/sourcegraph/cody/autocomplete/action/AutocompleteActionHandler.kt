package com.sourcegraph.cody.autocomplete.action

import com.intellij.openapi.actionSystem.DataContext
import com.intellij.openapi.diagnostic.Logger
import com.intellij.openapi.editor.Caret
import com.intellij.openapi.editor.Editor
import com.intellij.openapi.editor.actionSystem.EditorActionHandler
import com.intellij.openapi.project.Project
import com.sourcegraph.cody.agent.CodyAgent
import com.sourcegraph.cody.autocomplete.render.CodyAutocompleteElementRenderer
import com.sourcegraph.cody.autocomplete.render.InlayModelUtil
import com.sourcegraph.cody.vscode.InlineAutocompleteItem
import com.sourcegraph.utils.CodyEditorUtil

open class AutocompleteActionHandler : EditorActionHandler() {
  private val logger = Logger.getInstance(AutocompleteActionHandler::class.java)

  override fun isEnabledForCaret(editor: Editor, caret: Caret, dataContext: DataContext?): Boolean {
    // Returns false to fall back to normal action if there is no suggestion at the caret.
    val project = editor.project ?: return false
    return CodyEditorUtil.isEditorInstanceSupported(editor) &&
        hasAnyAutocompleteItems(project, caret)
  }

  private fun hasAnyAutocompleteItems(project: Project, caret: Caret): Boolean =
      CodyAgent.isConnected(project) && getCurrentAutocompleteItem(caret) != null

  private fun getAutocompleteRenderers(caret: Caret): List<CodyAutocompleteElementRenderer> =
      InlayModelUtil.getAllInlaysForEditor(caret.editor)
          .map { it.renderer }
          .filterIsInstance<CodyAutocompleteElementRenderer>()

  /**
   * Returns the autocompletion item for the first inlay of type `CodyAutocompleteElementRenderer`
   * regardless if the inlay is positioned at the caret. The reason we don't require the inlay to be
   * positioned at the caret is that completions can suggest changes in a nearby character like in
   * this situation:
   *
   * ` System.out.println("a: CARET"); // original System.out.println("a: " + a);CARET //
   * autocomplete ` *
   */
  protected fun getCurrentAutocompleteItem(caret: Caret): InlineAutocompleteItem? =
      getAutocompleteRenderers(caret).firstNotNullOfOrNull { it.completionItems.firstOrNull() }

  protected fun getAllAutocompleteItems(caret: Caret): List<InlineAutocompleteItem> =
      getAutocompleteRenderers(caret).flatMap { it.completionItems }.distinct()

  protected fun getSingleCaret(editor: Editor): Caret? {
    val allCarets = editor.caretModel.allCarets
    // Only accept completions if there's a single caret.
    return if (allCarets.size < 2) allCarets.firstOrNull() else null
  }
}
