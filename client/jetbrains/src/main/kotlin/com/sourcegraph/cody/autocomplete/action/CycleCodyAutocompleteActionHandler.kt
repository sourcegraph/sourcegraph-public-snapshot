package com.sourcegraph.cody.autocomplete.action

import com.intellij.openapi.actionSystem.DataContext
import com.intellij.openapi.diagnostic.Logger
import com.intellij.openapi.editor.Caret
import com.intellij.openapi.editor.Editor
import com.sourcegraph.cody.agent.CodyAgent
import com.sourcegraph.cody.vscode.InlineAutocompleteItem
import com.sourcegraph.utils.CodyEditorUtil
import java.util.concurrent.ConcurrentHashMap

class CycleCodyAutocompleteActionHandler(private val cycleDirection: CycleDirection) :
    AutocompleteActionHandler() {
  private val logger = Logger.getInstance(CycleCodyAutocompleteActionHandler::class.java)

  override fun isEnabledForCaret(editor: Editor, caret: Caret, dataContext: DataContext?): Boolean {
    val project = editor.project ?: return false
    val allAutocompleteItems = getAllAutocompleteItems(caret)
    autocompleteItemsCache[editor.cycleAutocompleteCacheKey(caret)] = allAutocompleteItems
    return CodyEditorUtil.isEditorInstanceSupported(editor) &&
        CodyAgent.isConnected(project) &&
        allAutocompleteItems.isNotEmpty()
  }

  override fun doExecute(editor: Editor, maybeCaret: Caret?, dataContext: DataContext?) {
    (maybeCaret ?: getSingleCaret(editor) ?: return).let { caret ->
      val cacheKey = editor.cycleAutocompleteCacheKey(caret)
      val allItems = autocompleteItemsCache[cacheKey] ?: emptyList()
      logger.warn("${cycleDirection.name} cycle trigger: ${allItems.size}")
      autocompleteItemsCache.remove(cacheKey)
    }
  }

  companion object {
    enum class CycleDirection {
      FORWARD,
      BACKWARD
    }

    class CacheKey(val caretOffset: Int, val documentName: String) {
      constructor(
          caret: Caret,
          editor: Editor
      ) : this(caret.offset, CodyEditorUtil.getVirtualFile(editor)?.name ?: "")

      override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other !is CacheKey) return false

        if (caretOffset != other.caretOffset) return false
        if (documentName != other.documentName) return false

        return true
      }

      override fun hashCode(): Int {
        var result = caretOffset
        result = 31 * result + documentName.hashCode()
        return result
      }
    }

    infix fun Editor.cycleAutocompleteCacheKey(caret: Caret) = CacheKey(caret, this)

    private val autocompleteItemsCache = ConcurrentHashMap<CacheKey, List<InlineAutocompleteItem>>()
  }
}
